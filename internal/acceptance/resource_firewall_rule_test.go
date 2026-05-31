package acceptance

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFirewallRule_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckFirewallRuleDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleConfigBasic("tf-acc-test-rule"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_firewall_rule.test", "name", "tf-acc-test-rule"),
					resource.TestCheckResourceAttr("openwrt_firewall_rule.test", "target", "ACCEPT"),
				),
			},
			{
				ResourceName:      "openwrt_firewall_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallRule_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckFirewallRuleDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleConfigBasic("tf-acc-test-rule"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_firewall_rule.test", "name", "tf-acc-test-rule"),
					resource.TestCheckResourceAttr("openwrt_firewall_rule.test", "target", "ACCEPT"),
				),
			},
			{
				Config: testAccFirewallRuleConfigUpdate("tf-acc-test-rule"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_firewall_rule.test", "name", "tf-acc-test-rule"),
					resource.TestCheckResourceAttr("openwrt_firewall_rule.test", "target", "REJECT"),
					resource.TestCheckResourceAttr("openwrt_firewall_rule.test", "dest_port", "8080"),
				),
			},
		},
	})
}

func testAccCheckFirewallRuleDestroyed(s *terraform.State) error {
	client := GetTestProvider(&testing.T{})
	if client == nil {
		return nil
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openwrt_firewall_rule" {
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

func testAccFirewallRuleConfigBasic(name string) string {
	return ProviderConfig() + `
resource "openwrt_firewall_rule" "test" {
  name   = "` + name + `"
  target = "ACCEPT"
}
`
}

func testAccFirewallRuleConfigUpdate(name string) string {
	return ProviderConfig() + `
resource "openwrt_firewall_rule" "test" {
  name       = "` + name + `"
  target     = "REJECT"
  proto      = "tcp"
  dest_port  = "8080"
}
`
}

func TestAccFirewallForwarding_basic(t *testing.T) {
	srcZone := GetFirewallSrcZone()
	destZone := GetFirewallDestZone()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckFirewallForwardingDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallForwardingConfigBasic(srcZone, destZone),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_firewall_forwarding.test", "src", srcZone),
					resource.TestCheckResourceAttr("openwrt_firewall_forwarding.test", "dest", destZone),
				),
			},
			{
				ResourceName:      "openwrt_firewall_forwarding.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckFirewallForwardingDestroyed(s *terraform.State) error {
	client := GetTestProvider(&testing.T{})
	if client == nil {
		return nil
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openwrt_firewall_forwarding" {
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

func testAccFirewallForwardingConfigBasic(src, dest string) string {
	return ProviderConfig() + `
resource "openwrt_firewall_forwarding" "test" {
  src  = "` + src + `"
  dest = "` + dest + `"
}
`
}
