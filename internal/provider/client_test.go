package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientAuthAndSysCall(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/luci/rpc/auth", func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode auth request: %v", err)
		}
		if req.Method != "login" {
			t.Fatalf("expected login, got %s", req.Method)
		}

		resp := rpcResponse{
			ID:     req.ID,
			Result: json.RawMessage(`"testtoken"`),
			Error:  nil,
		}
		_ = json.NewEncoder(w).Encode(&resp)
	})

	mux.HandleFunc("/cgi-bin/luci/rpc/sys", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("auth"); got != "testtoken" {
			t.Fatalf("expected auth token testtoken, got %q", got)
		}
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode sys request: %v", err)
		}
		if req.Method != "hostname" {
			t.Fatalf("expected hostname, got %s", req.Method)
		}

		resp := rpcResponse{
			ID:     req.ID,
			Result: json.RawMessage(`"openwrt-test"`),
			Error:  nil,
		}
		_ = json.NewEncoder(w).Encode(&resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	ctx := context.Background()
	client, err := NewJsonRpcClient(ctx, JsonRpcConfig{
		BaseURL:  srv.URL,
		Username: "root",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	raw, err := client.SysCall(ctx, "hostname")
	if err != nil {
		t.Fatalf("SysCall failed: %v", err)
	}

	var hostname string
	if err := json.Unmarshal(raw, &hostname); err != nil {
		t.Fatalf("failed to unmarshal hostname: %v", err)
	}
	if hostname != "openwrt-test" {
		t.Fatalf("expected hostname openwrt-test, got %q", hostname)
	}
}

func TestNewClientRequiresBaseURL(t *testing.T) {
	_, err := NewJsonRpcClient(context.Background(), JsonRpcConfig{})
	if err == nil {
		t.Fatalf("expected error for empty BaseURL")
	}
}

func TestUCISection_StringResponse(t *testing.T) {
	mux := http.NewServeMux()
	setupAuthHandler(mux)
	mux.HandleFunc("/cgi-bin/luci/rpc/uci", func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode uci request: %v", err)
		}
		if req.Method != "section" {
			t.Fatalf("expected section, got %s", req.Method)
		}
		resp := rpcResponse{
			ID:     req.ID,
			Result: json.RawMessage(`"lan"`),
			Error:  nil,
		}
		_ = json.NewEncoder(w).Encode(&resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ctx := context.Background()

	secName, err := client.UCISection(ctx, "network", "interface", "lan", map[string]interface{}{
		"proto":  "static",
		"ifname": "eth0",
	})
	if err != nil {
		t.Fatalf("UCISection failed: %v", err)
	}
	if secName != "lan" {
		t.Fatalf("expected section name 'lan', got %q", secName)
	}
}

func TestUCISection_BoolTrueResponse(t *testing.T) {
	mux := http.NewServeMux()
	setupAuthHandler(mux)
	mux.HandleFunc("/cgi-bin/luci/rpc/uci", func(w http.ResponseWriter, r *http.Request) {
		resp := rpcResponse{
			ID:     1,
			Result: json.RawMessage(`true`),
			Error:  nil,
		}
		_ = json.NewEncoder(w).Encode(&resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ctx := context.Background()

	secName, err := client.UCISection(ctx, "network", "interface", "lan", map[string]interface{}{
		"proto": "static",
	})
	if err != nil {
		t.Fatalf("UCISection failed: %v", err)
	}
	if secName != "lan" {
		t.Fatalf("expected section name 'lan' from bool true, got %q", secName)
	}
}

func TestUCISection_BoolFalseResponse(t *testing.T) {
	mux := http.NewServeMux()
	setupAuthHandler(mux)
	mux.HandleFunc("/cgi-bin/luci/rpc/uci", func(w http.ResponseWriter, r *http.Request) {
		resp := rpcResponse{
			ID:     1,
			Result: json.RawMessage(`false`),
			Error:  nil,
		}
		_ = json.NewEncoder(w).Encode(&resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ctx := context.Background()

	_, err := client.UCISection(ctx, "network", "interface", "lan", nil)
	if err == nil {
		t.Fatal("expected error for bool false response")
	}
}

func TestUCIGetAll(t *testing.T) {
	mux := http.NewServeMux()
	setupAuthHandler(mux)
	mux.HandleFunc("/cgi-bin/luci/rpc/uci", func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode uci request: %v", err)
		}
		if req.Method != "get_all" {
			t.Fatalf("expected get_all, got %s", req.Method)
		}
		resp := rpcResponse{
			ID:     req.ID,
			Result: json.RawMessage(`{"interface":[{"ifname":"eth0","proto":"static"}]}`),
			Error:  nil,
		}
		_ = json.NewEncoder(w).Encode(&resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ctx := context.Background()

	result, err := client.UCIGetAll(ctx, "network", "")
	if err != nil {
		t.Fatalf("UCIGetAll failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestUCIGetAll_WithSection(t *testing.T) {
	mux := http.NewServeMux()
	setupAuthHandler(mux)
	mux.HandleFunc("/cgi-bin/luci/rpc/uci", func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode uci request: %v", err)
		}
		resp := rpcResponse{
			ID:     req.ID,
			Result: json.RawMessage(`{".type":"interface",".name":"lan","ifname":"eth0","proto":"static","ipaddr":"192.168.2.1/24"}`),
			Error:  nil,
		}
		_ = json.NewEncoder(w).Encode(&resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ctx := context.Background()

	result, err := client.UCIGetAll(ctx, "network", "lan")
	if err != nil {
		t.Fatalf("UCIGetAll with section failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result["ifname"] != "eth0" {
		t.Fatalf("expected ifname 'eth0', got %v", result["ifname"])
	}
}

func TestUCIDelete(t *testing.T) {
	mux := http.NewServeMux()
	setupAuthHandler(mux)
	mux.HandleFunc("/cgi-bin/luci/rpc/uci", func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode uci request: %v", err)
		}
		if req.Method != "delete" {
			t.Fatalf("expected delete, got %s", req.Method)
		}
		resp := rpcResponse{
			ID:     req.ID,
			Result: json.RawMessage(`true`),
			Error:  nil,
		}
		_ = json.NewEncoder(w).Encode(&resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ctx := context.Background()

	err := client.UCIDelete(ctx, "network", "lan")
	if err != nil {
		t.Fatalf("UCIDelete failed: %v", err)
	}
}

func TestUCICommit(t *testing.T) {
	mux := http.NewServeMux()
	setupAuthHandler(mux)
	mux.HandleFunc("/cgi-bin/luci/rpc/uci", func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode uci request: %v", err)
		}
		if req.Method != "commit" {
			t.Fatalf("expected commit, got %s", req.Method)
		}
		resp := rpcResponse{
			ID:     req.ID,
			Result: json.RawMessage(`true`),
			Error:  nil,
		}
		_ = json.NewEncoder(w).Encode(&resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ctx := context.Background()

	err := client.UCICommit(ctx, "network")
	if err != nil {
		t.Fatalf("UCICommit failed: %v", err)
	}
}

func TestUCIApply(t *testing.T) {
	mux := http.NewServeMux()
	setupAuthHandler(mux)
	mux.HandleFunc("/cgi-bin/luci/rpc/uci", func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode uci request: %v", err)
		}
		if req.Method != "apply" {
			t.Fatalf("expected apply, got %s", req.Method)
		}
		resp := rpcResponse{
			ID:     req.ID,
			Result: json.RawMessage(`true`),
			Error:  nil,
		}
		_ = json.NewEncoder(w).Encode(&resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ctx := context.Background()

	err := client.UCIApply(ctx, false)
	if err != nil {
		t.Fatalf("UCIApply failed: %v", err)
	}
}

func TestUCITSet(t *testing.T) {
	mux := http.NewServeMux()
	setupAuthHandler(mux)
	mux.HandleFunc("/cgi-bin/luci/rpc/uci", func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode uci request: %v", err)
		}
		if req.Method != "tset" {
			t.Fatalf("expected tset, got %s", req.Method)
		}
		resp := rpcResponse{
			ID:     req.ID,
			Result: json.RawMessage(`true`),
			Error:  nil,
		}
		_ = json.NewEncoder(w).Encode(&resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ctx := context.Background()

	err := client.UCITSet(ctx, "network", "lan", map[string]interface{}{
		"ipaddr": "192.168.2.1",
	})
	if err != nil {
		t.Fatalf("UCITSet failed: %v", err)
	}
}

func TestUCISection_WithBoolValues(t *testing.T) {
	mux := http.NewServeMux()
	setupAuthHandler(mux)
	mux.HandleFunc("/cgi-bin/luci/rpc/uci", func(w http.ResponseWriter, r *http.Request) {
		resp := rpcResponse{
			ID:     1,
			Result: json.RawMessage(`"lan"`),
			Error:  nil,
		}
		_ = json.NewEncoder(w).Encode(&resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ctx := context.Background()

	_, err := client.UCISection(ctx, "network", "interface", "lan", map[string]interface{}{
		"proto":    "static",
		"delegate": true,
		"metric":   0,
	})
	if err != nil {
		t.Fatalf("UCISection with mixed values failed: %v", err)
	}
}

func setupAuthHandler(mux *http.ServeMux) {
	mux.HandleFunc("/cgi-bin/luci/rpc/auth", func(w http.ResponseWriter, r *http.Request) {
		resp := rpcResponse{
			ID:     1,
			Result: json.RawMessage(`"testtoken"`),
			Error:  nil,
		}
		_ = json.NewEncoder(w).Encode(&resp)
	})
}

func newTestClient(t *testing.T, baseURL string) *JsonRpcClient {
	t.Helper()
	client, err := NewJsonRpcClient(context.Background(), JsonRpcConfig{
		BaseURL:  baseURL,
		Username: "root",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	return client
}
