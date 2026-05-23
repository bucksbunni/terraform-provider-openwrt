package acceptance

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFirewallZone_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckFirewallZoneDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallZoneConfigBasic("tf-acc-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_firewall_zone.test", "name", "tf-acc-test"),
					resource.TestCheckResourceAttr("openwrt_firewall_zone.test", "input", "ACCEPT"),
					resource.TestCheckResourceAttr("openwrt_firewall_zone.test", "output", "ACCEPT"),
					resource.TestCheckResourceAttr("openwrt_firewall_zone.test", "forward", "ACCEPT"),
				),
			},
			{
				ResourceName:      "openwrt_firewall_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallZone_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckFirewallZoneDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallZoneConfigBasic("tf-acc-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_firewall_zone.test", "name", "tf-acc-test"),
					resource.TestCheckResourceAttr("openwrt_firewall_zone.test", "input", "ACCEPT"),
				),
			},
			{
				Config: testAccFirewallZoneConfigUpdate("tf-acc-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_firewall_zone.test", "name", "tf-acc-test"),
					resource.TestCheckResourceAttr("openwrt_firewall_zone.test", "input", "DROP"),
					resource.TestCheckResourceAttr("openwrt_firewall_zone.test", "output", "DROP"),
					resource.TestCheckResourceAttr("openwrt_firewall_zone.test", "forward", "REJECT"),
					resource.TestCheckResourceAttr("openwrt_firewall_zone.test", "masq", "true"),
				),
			},
		},
	})
}

func testAccCheckFirewallZoneDestroyed(s *terraform.State) error {
	client := GetTestProvider(&testing.T{})
	if client == nil {
		return nil
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openwrt_firewall_zone" {
			continue
		}

		if rs.Primary.ID == "" {
			continue
		}

		parts := splitImportID(rs.Primary.ID)
		if len(parts) != 2 {
			continue
		}

		data, err := client.UCIGetAll(context.Background(), parts[0], parts[1])
		if err != nil {
			return nil
		}

		if len(data) > 0 {
			return nil
		}
	}

	return nil
}

func testAccFirewallZoneConfigBasic(name string) string {
	return ProviderConfig() + `
resource "openwrt_firewall_zone" "test" {
  name   = "` + name + `"
  input  = "ACCEPT"
  output = "ACCEPT"
  forward = "ACCEPT"
}
`
}

func testAccFirewallZoneConfigUpdate(name string) string {
	return ProviderConfig() + `
resource "openwrt_firewall_zone" "test" {
  name   = "` + name + `"
  input  = "DROP"
  output = "DROP"
  forward = "REJECT"
  masq   = true
}
`
}
