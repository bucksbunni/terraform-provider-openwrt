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

var optionalEnvVars = []string{
	"OPENWRT_INSECURE",
	"OPENWRT_SKIP_PRECHECK",
	"OPENWRT_SKIP_CONNECTIVITY",
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

func RandomWithPrefix(prefix string) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return prefix + "-" + string(b)
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
