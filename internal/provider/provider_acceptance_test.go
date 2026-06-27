//go:build acceptance
// +build acceptance

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccProtoV6ProviderFactories wires the provider into the plugin-testing
// framework. The factory must yield a tfprotov6.ProviderServer, which
// providerserver.NewProtocol6WithError builds from the framework provider.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"openwrt": providerserver.NewProtocol6WithError(NewProvider("test")()),
}

// TestAccSysRPCHostname runs against a real OpenWrt (or a richer test HTTP
// server). It is gated by OPENWRT_HOST and, like all resource.Test cases, only
// runs when TF_ACC is set. Build with -tags acceptance to include it.
func TestAccSysRPCHostname(t *testing.T) {
	host := os.Getenv("OPENWRT_HOST")
	if host == "" {
		t.Skip("OPENWRT_HOST not set")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "openwrt" {
  host     = "` + host + `"
  username = "` + os.Getenv("OPENWRT_USERNAME") + `"
  password = "` + os.Getenv("OPENWRT_PASSWORD") + `"
}

data "openwrt_sys_rpc" "hostname" {
  method = "hostname"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.openwrt_sys_rpc.hostname", "result_json"),
				),
			},
		},
	})
}
