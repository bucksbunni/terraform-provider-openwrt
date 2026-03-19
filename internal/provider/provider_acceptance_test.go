//go:build acceptance
// +build acceptance

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Runs against a real OpenWrt (or a richer test HTTP server)
func TestAccSysRPCHostname(t *testing.T) {
	host := os.Getenv("OPENWRT_HOST")
	if host == "" {
		t.Skip("OPENWRT_HOST not set")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (provider.Provider, error){
			"openwrt": func() (provider.Provider, error) {
				return New("test")(), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: `
provider "openwrt" {
  host     = "` + host + `"
  username = "` + os.Getenv("OPENWRT_USER") + `"
  password = "` + os.Getenv("OPENWRT_PASS") + `"
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
