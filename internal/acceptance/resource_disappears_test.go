package acceptance

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFirewallZone_Disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckFirewallZoneDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallZoneConfigBasic("tf-acc-disappear"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_firewall_zone.test", "name", "tf-acc-disappear"),
					testAccCheckResourceDisappears("firewall", "zone", "tf-acc-disappear"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNetworkInterface_Disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckNetworkInterfaceDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkInterfaceConfigBasic("tf_acc_disappear", "static"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_interface.test", "name", "tf_acc_disappear"),
					testAccCheckResourceDisappears("network", "interface", "tf_acc_disappear"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDHCPPool_Disappears(t *testing.T) {
	dhcpInterface := GetDHCPInterface()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPPoolDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPPoolConfigBasic("tf-acc-disappear", dhcpInterface),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_dhcp_pool.test", "interface", dhcpInterface),
					testAccCheckDHCPPoolDisappears("tf-acc-disappear"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckResourceDisappears(config, sectionType, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := GetTestProvider(&testing.T{})
		if client == nil {
			return nil
		}

		sections, err := client.UCIForeach(context.Background(), config, sectionType)
		if err != nil {
			return err
		}

		var secName string
		for _, sec := range sections {
			// Match either the "name" option (anonymous sections such as
			// firewall zones) or the section identifier itself (named sections
			// such as network interfaces).
			if sec["name"] == name || sec[".name"] == name {
				secName = sec[".name"].(string)
				break
			}
		}

		if secName == "" {
			return nil
		}

		err = client.UCIDelete(context.Background(), config, secName)
		if err != nil {
			return err
		}

		err = client.UCICommit(context.Background(), config)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckDHCPPoolDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := GetTestProvider(&testing.T{})
		if client == nil {
			return nil
		}

		entries, err := client.UCIForeach(context.Background(), "dhcp", "dhcp")
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if entryName, ok := entry["name"].(string); ok && entryName == name {
				if sectionName, ok := entry[".name"].(string); ok {
					err := client.UCIDelete(context.Background(), "dhcp", sectionName)
					if err != nil {
						return err
					}
					return client.UCICommit(context.Background(), "dhcp")
				}
			}
		}

		return nil
	}
}
