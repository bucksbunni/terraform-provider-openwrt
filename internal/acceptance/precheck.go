package acceptance

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/bucksbunni/terraform-provider-openwrt/internal/provider"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"openwrt": providerserver.NewProtocol6WithError(provider.NewProvider("acctest")()),
}

func TestAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return testAccProtoV6ProviderFactories
}

func ProviderConfig() string {
	host := os.Getenv("OPENWRT_HOST")
	username := os.Getenv("OPENWRT_USERNAME")
	password := os.Getenv("OPENWRT_PASSWORD")
	insecure := os.Getenv("OPENWRT_INSECURE")

	insecureStr := "false"
	if insecure == "true" {
		insecureStr = "true"
	}

	return `
provider "openwrt" {
  host     = "` + host + `"
  username = "` + username + `"
  password = "` + password + `"
  insecure = ` + insecureStr + `
}
`
}

var requiredEnvVars = []string{
	"OPENWRT_HOST",
	"OPENWRT_USERNAME",
	"OPENWRT_PASSWORD",
}

func PreCheck(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance tests skipped unless TF_ACC=1 is set")
	}

	if os.Getenv("OPENWRT_SKIP_PRECHECK") == "true" {
		return
	}

	missing := make([]string, 0, len(requiredEnvVars))
	for _, v := range requiredEnvVars {
		if os.Getenv(v) == "" {
			missing = append(missing, v)
		}
	}

	if len(missing) > 0 {
		t.Skipf("acceptance test skipped: environment variables not set: %v", missing)
	}
}

func PreCheckWithConnectivity(t *testing.T) {
	PreCheck(t)

	if os.Getenv("OPENWRT_SKIP_CONNECTIVITY") == "true" {
		return
	}

	host := GetTestHost()
	username := GetTestUsername()
	password := GetTestPassword()
	insecure := GetTestInsecure()

	client, err := provider.NewJsonRpcClient(t.Context(), provider.JsonRpcConfig{
		BaseURL:  host,
		Username: username,
		Password: password,
		Insecure: insecure,
	})
	if err != nil {
		t.Skipf("acceptance test skipped: failed to create client: %v", err)
	}

	_, err = client.SysCall(t.Context(), "hostname")
	if err != nil {
		t.Skipf("acceptance test skipped: failed to connect to OpenWrt host: %v", err)
	}
}

// RequireWireless skips the test unless the target device exposes at least one
// mac80211 radio. Wireless resources need wireless packages and radio hardware,
// which are absent on wired routers and minimal VMs.
func RequireWireless(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		return
	}
	client := GetTestProvider(t)
	out, err := client.SysExec(t.Context(), "ls /sys/class/ieee80211 2>/dev/null")
	if err != nil || strings.TrimSpace(out) == "" {
		t.Skip("skipping wireless test: no mac80211 radio present on the device")
	}
}

// RequireWireguard skips the test unless the WireGuard kernel module is loaded.
// WireGuard resources require the kernel module and userspace tools.
func RequireWireguard(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		return
	}
	client := GetTestProvider(t)
	out, err := client.SysExec(t.Context(), "[ -d /sys/module/wireguard ] && echo yes || true")
	if err != nil || strings.TrimSpace(out) != "yes" {
		t.Skip("skipping WireGuard test: the wireguard kernel module is not loaded")
	}
}

func GetTestHost() string {
	return os.Getenv("OPENWRT_HOST")
}

func GetTestUsername() string {
	return os.Getenv("OPENWRT_USERNAME")
}

func GetTestPassword() string {
	return os.Getenv("OPENWRT_PASSWORD")
}

func GetTestInsecure() bool {
	return os.Getenv("OPENWRT_INSECURE") == "true"
}

func GetTestProvider(t *testing.T) *provider.JsonRpcClient {
	host := GetTestHost()
	username := GetTestUsername()
	password := GetTestPassword()
	insecure := GetTestInsecure()

	client, err := provider.NewJsonRpcClient(context.Background(), provider.JsonRpcConfig{
		BaseURL:  host,
		Username: username,
		Password: password,
		Insecure: insecure,
	})
	if err != nil {
		t.Fatalf("failed to create test provider: %v", err)
	}

	return client
}

func splitImportID(id string) []string {
	var parts []string
	for _, p := range strings.FieldsFunc(id, func(r rune) bool {
		return r == '.' || r == '/'
	}) {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func CheckResourceDestroyed(resourceName, config, sectionType, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return nil
		}

		if rs.Primary.ID == "" {
			return nil
		}

		client := GetTestProvider(&testing.T{})
		if client == nil {
			return nil
		}

		_, err := client.UCIGetAll(context.Background(), config, name)
		if err != nil {
			return nil
		}

		return nil
	}
}
