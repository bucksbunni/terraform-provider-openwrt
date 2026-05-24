package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// JsonRpcConfig holds the configuration for creating a JSON-RPC client.
type JsonRpcConfig struct {
	// BaseURL is the URL of the LuCI RPC endpoint (e.g., http://192.168.1.1).
	BaseURL string
	// Username is the OpenWrt login username (typically 'root').
	Username string
	// Password is the OpenWrt login password.
	Password string
	// Insecure controls TLS certificate verification for HTTPS connections.
	Insecure bool
}

// JsonRpcClient provides access to the OpenWrt LuCI JSON-RPC API.
// It handles authentication and provides methods for UCI, filesystem, and system operations.
type JsonRpcClient struct {
	baseURL  *url.URL
	username string
	password string

	httpClient *http.Client

	mu    sync.Mutex
	token string
	id    int64
}

// NewJsonRpcClient creates a new JSON-RPC client for communicating with OpenWrt's LuCI API.
// It validates the configuration and sets up the HTTP client.
func NewJsonRpcClient(ctx context.Context, cfg JsonRpcConfig) (*JsonRpcClient, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("BaseURL is required")
	}

	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	transport := &http.Transport{}

	if u.Scheme == "https" && cfg.Insecure {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
		}
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return &JsonRpcClient{
		baseURL:    u,
		username:   cfg.Username,
		password:   cfg.Password,
		httpClient: httpClient,
	}, nil
}

type rpcRequest struct {
	ID     int64         `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	ID     int64           `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error"`
}

// ensureAuth ensures we have a valid auth token.
func (c *JsonRpcClient) ensureAuth(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" {
		return nil
	}

	// POST /cgi-bin/luci/rpc/auth
	endpoint := *c.baseURL
	endpoint.Path = "/cgi-bin/luci/rpc/auth"

	reqBody := rpcRequest{
		ID:     1,
		Method: "login",
		Params: []interface{}{c.username, c.password},
	}

	buf, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal auth request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("create auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth HTTP error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth failed: HTTP %d", resp.StatusCode)
	}

	var rpcResp rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return fmt.Errorf("decode auth response: %w", err)
	}

	if rpcResp.Error != nil {
		return fmt.Errorf("auth RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	var token string
	if err := json.Unmarshal(rpcResp.Result, &token); err != nil {
		return fmt.Errorf("decode auth token: %w", err)
	}

	if token == "" {
		return fmt.Errorf("empty auth token")
	}

	c.token = token
	return nil
}

// call performs a JSON-RPC call on a given library, e.g. "uci", "fs", "sys", "ipkg".
func (c *JsonRpcClient) call(ctx context.Context, library, method string, params ...interface{}) (json.RawMessage, error) {
	if err := c.ensureAuth(ctx); err != nil {
		return nil, err
	}

	return c.callWithToken(ctx, library, method, params...)
}

func (c *JsonRpcClient) nextID() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.id++
	return c.id
}

func (c *JsonRpcClient) callWithToken(ctx context.Context, library, method string, params ...interface{}) (json.RawMessage, error) {
	endpoint := *c.baseURL
	endpoint.Path = "/cgi-bin/luci/rpc/" + library

	q := endpoint.Query()
	q.Set("auth", c.token)
	endpoint.RawQuery = q.Encode()

	reqBody := rpcRequest{
		ID:     c.nextID(),
		Method: method,
		Params: params,
	}

	buf, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal RPC request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("create RPC request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("RPC HTTP error: %w", err)
	}
	defer resp.Body.Close()

	// If 403, token might be invalid; retry once with fresh auth.
	if resp.StatusCode == http.StatusForbidden {
		c.mu.Lock()
		c.token = ""
		c.mu.Unlock()

		if err := c.ensureAuth(ctx); err != nil {
			return nil, err
		}
		return c.callWithToken(ctx, library, method, params...)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RPC failed: HTTP %d", resp.StatusCode)
	}

	var rpcResp rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("decode RPC response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

// Convenience helpers for specific libraries:

// UCI

func (c *JsonRpcClient) UCIGetAll(ctx context.Context, config, section string) (map[string]interface{}, error) {
	params := []interface{}{config}
	if section != "" {
		params = append(params, section)
	}
	raw, err := c.call(ctx, "uci", "get_all", params...)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode UCI get_all: %w", err)
	}
	return result, nil
}

func (c *JsonRpcClient) UCISection(ctx context.Context, config, typ, name string, values map[string]interface{}) (string, error) {
	// section(config, type, name, values) -> section name or true/false
	raw, err := c.call(ctx, "uci", "section", config, typ, name, values)
	if err != nil {
		return "", err
	}

	var secName string
	if err := json.Unmarshal(raw, &secName); err != nil {
		var boolResult bool
		if err := json.Unmarshal(raw, &boolResult); err != nil {
			return "", fmt.Errorf("decode UCI section result: %w", err)
		}
		if boolResult {
			return name, nil
		}
		return "", fmt.Errorf("UCI section creation failed for %s.%s", config, name)
	}
	return secName, nil
}

func (c *JsonRpcClient) UCIDelete(ctx context.Context, config, section string) error {
	_, err := c.call(ctx, "uci", "delete", config, section)
	return err
}

func (c *JsonRpcClient) UCICommit(ctx context.Context, config string) error {
	_, err := c.call(ctx, "uci", "commit", config)
	return err
}

func (c *JsonRpcClient) UCIApply(ctx context.Context, rollback bool) error {
	_, err := c.call(ctx, "uci", "apply", rollback)
	return err
}

func (c *JsonRpcClient) UCITSet(ctx context.Context, config, section string, values map[string]interface{}) error {
	_, err := c.call(ctx, "uci", "tset", config, section, values)
	return err
}

// FS

func (c *JsonRpcClient) FSReadFile(ctx context.Context, path string) (string, error) {
	raw, err := c.call(ctx, "fs", "readfile", path)
	if err != nil {
		return "", err
	}
	var b64 string
	if err := json.Unmarshal(raw, &b64); err != nil {
		return "", fmt.Errorf("decode fs.readfile result: %w", err)
	}
	return b64, nil
}

func (c *JsonRpcClient) FSWriteFile(ctx context.Context, path, base64Content string) error {
	_, err := c.call(ctx, "fs", "writefile", path, base64Content)
	return err
}

func (c *JsonRpcClient) FSUnlink(ctx context.Context, path string) error {
	_, err := c.call(ctx, "fs", "unlink", path)
	return err
}

func (c *JsonRpcClient) FSStat(ctx context.Context, path string) (map[string]interface{}, error) {
	raw, err := c.call(ctx, "fs", "stat", path)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode fs.stat: %w", err)
	}
	return result, nil
}

// IPKG

func (c *JsonRpcClient) IPKGInstalled(ctx context.Context, pkg string) (bool, error) {
	raw, err := c.call(ctx, "ipkg", "installed", pkg)
	if err != nil {
		return false, err
	}
	var ok bool
	if err := json.Unmarshal(raw, &ok); err != nil {
		return false, fmt.Errorf("decode ipkg.installed: %w", err)
	}
	return ok, nil
}

func (c *JsonRpcClient) IPKGInstall(ctx context.Context, pkg string) error {
	_, err := c.call(ctx, "ipkg", "install", pkg)
	return err
}

func (c *JsonRpcClient) IPKGRemove(ctx context.Context, pkg string) error {
	_, err := c.call(ctx, "ipkg", "remove", pkg)
	return err
}

// IPKGListAll lists all packages known to opkg, optionally matching a pattern.
func (c *JsonRpcClient) IPKGListAll(ctx context.Context, pattern string) ([]map[string]interface{}, error) {
	params := []interface{}{}
	if pattern != "" {
		params = append(params, pattern)
	}
	raw, err := c.call(ctx, "ipkg", "list_all", params...)
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode ipkg.list_all: %w", err)
	}
	return result, nil
}

// IPKGListInstalled lists installed packages, optionally matching a pattern.
func (c *JsonRpcClient) IPKGListInstalled(ctx context.Context, pattern string) ([]map[string]interface{}, error) {
	params := []interface{}{}
	if pattern != "" {
		params = append(params, pattern)
	}
	raw, err := c.call(ctx, "ipkg", "list_installed", params...)
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode ipkg.list_installed: %w", err)
	}
	return result, nil
}

// IPKGFind finds packages whose name or description matches the pattern.
func (c *JsonRpcClient) IPKGFind(ctx context.Context, pattern string) ([]map[string]interface{}, error) {
	raw, err := c.call(ctx, "ipkg", "find", pattern)
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode ipkg.find: %w", err)
	}
	return result, nil
}

// IPKGInfo returns information about installed and available packages.
// If pkg is empty, info for all packages is returned.
func (c *JsonRpcClient) IPKGInfo(ctx context.Context, pkg string) (map[string]interface{}, error) {
	params := []interface{}{}
	if pkg != "" {
		params = append(params, pkg)
	}
	raw, err := c.call(ctx, "ipkg", "info", params...)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode ipkg.info: %w", err)
	}
	return result, nil
}

// IPKGStatus returns the package status for the given package (or all if empty).
func (c *JsonRpcClient) IPKGStatus(ctx context.Context, pkg string) (map[string]interface{}, error) {
	params := []interface{}{}
	if pkg != "" {
		params = append(params, pkg)
	}
	raw, err := c.call(ctx, "ipkg", "status", params...)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode ipkg.status: %w", err)
	}
	return result, nil
}

// IPKGOverlayRoot returns the overlay root used by opkg.
func (c *JsonRpcClient) IPKGOverlayRoot(ctx context.Context) (string, error) {
	raw, err := c.call(ctx, "ipkg", "overlay_root")
	if err != nil {
		return "", err
	}
	var root string
	if err := json.Unmarshal(raw, &root); err != nil {
		return "", fmt.Errorf("decode ipkg.overlay_root: %w", err)
	}
	return root, nil
}

// IPKGCompareVersions compares two versions using the given operator
// ("<=", "<", ">", ">=", "=", "<<", ">>", "~=").
func (c *JsonRpcClient) IPKGCompareVersions(ctx context.Context, ver1, ver2, comp string) (bool, error) {
	raw, err := c.call(ctx, "ipkg", "compare_versions", ver1, ver2, comp)
	if err != nil {
		return false, err
	}
	var ok bool
	if err := json.Unmarshal(raw, &ok); err != nil {
		return false, fmt.Errorf("decode ipkg.compare_versions: %w", err)
	}
	return ok, nil
}

// IPKGUpdate runs "opkg update" via luci.model.ipkg.update().
func (c *JsonRpcClient) IPKGUpdate(ctx context.Context) error {
	_, err := c.call(ctx, "ipkg", "update")
	return err
}

// IPKGUpgrade runs "opkg upgrade" via luci.model.ipkg.upgrade().
func (c *JsonRpcClient) IPKGUpgrade(ctx context.Context) error {
	_, err := c.call(ctx, "ipkg", "upgrade")
	return err
}

// SysCall calls the /rpc/sys JSON-RPC endpoint with the given method.
// The method can be any luci.sys* function, e.g. "hostname", "uptime",
// "net.routes", "user.getuser", "process.list", "wifi.getiwinfo", etc.
func (c *JsonRpcClient) SysCall(ctx context.Context, method string, params ...interface{}) (json.RawMessage, error) {
	return c.call(ctx, "sys", method, params...)
}

// Optional UCI calls

// UCIAdd adds an anonymous section and returns its name.
func (c *JsonRpcClient) UCIAdd(ctx context.Context, config, typ string) (string, error) {
	raw, err := c.call(ctx, "uci", "add", config, typ)
	if err != nil {
		return "", err
	}
	var name string
	if err := json.Unmarshal(raw, &name); err != nil {
		return "", fmt.Errorf("decode UCI add: %w", err)
	}
	return name, nil
}

// UCISetSection creates or updates a named section (e.g., uci set network test interface).
func (c *JsonRpcClient) UCISetSection(ctx context.Context, config, section, typ string) error {
	_, err := c.call(ctx, "uci", "set", config, section, typ)
	return err
}

// UCIChanges returns the table of saved but uncommitted changes.
// If config is empty, changes for all configs are returned.
func (c *JsonRpcClient) UCIChanges(ctx context.Context, config string) (map[string]interface{}, error) {
	params := []interface{}{}
	if config != "" {
		params = append(params, config)
	}
	raw, err := c.call(ctx, "uci", "changes", params...)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode UCI changes: %w", err)
	}
	return result, nil
}

// UCIGet gets a section type or option value.
// If option is empty, the section type is returned.
func (c *JsonRpcClient) UCIGet(ctx context.Context, config, section, option string) (interface{}, error) {
	params := []interface{}{config, section}
	if option != "" {
		params = append(params, option)
	}
	raw, err := c.call(ctx, "uci", "get", params...)
	if err != nil {
		return nil, err
	}
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, fmt.Errorf("decode UCI get: %w", err)
	}
	return v, nil
}

// UCIGetState reads a value from the state directory.
// If option is empty, the section type is returned.
func (c *JsonRpcClient) UCIGetState(ctx context.Context, config, section, option string) (interface{}, error) {
	params := []interface{}{config, section}
	if option != "" {
		params = append(params, option)
	}
	raw, err := c.call(ctx, "uci", "get_state", params...)
	if err != nil {
		return nil, err
	}
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, fmt.Errorf("decode UCI get_state: %w", err)
	}
	return v, nil
}

// UCIRevert reverts saved but uncommitted changes for the given config.
func (c *JsonRpcClient) UCIRevert(ctx context.Context, config string) error {
	_, err := c.call(ctx, "uci", "revert", config)
	return err
}

// UCIDeleteAll deletes all sections of a given type (if type != "") or all
// sections in a config (if type == "").
func (c *JsonRpcClient) UCIDeleteAll(ctx context.Context, config, typ string) error {
	params := []interface{}{config}
	if typ != "" {
		params = append(params, typ)
	}
	_, err := c.call(ctx, "uci", "delete_all", params...)
	return err
}

// UCIForeach calls the callback for each section of a given type and returns
// the resulting list of section tables.
func (c *JsonRpcClient) UCIForeach(ctx context.Context, config, typ string) ([]map[string]interface{}, error) {
	params := []interface{}{config}
	if typ != "" {
		params = append(params, typ)
	}
	raw, err := c.call(ctx, "uci", "foreach", params...)
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode UCI foreach: %w", err)
	}
	return result, nil
}

// UCISet sets a value or creates a named section.
// See luci.model.uci.set() docs for details.
func (c *JsonRpcClient) UCISet(ctx context.Context, config, section, option string, value interface{}) error {
	_, err := c.call(ctx, "uci", "set", config, section, option, value)
	return err
}
