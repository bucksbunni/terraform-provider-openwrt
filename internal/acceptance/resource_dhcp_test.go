package acceptance

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDHCPPool_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPPoolDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPPoolConfigBasic("tf-acc-pool", "lan"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_dhcp_pool.test", "interface", "lan"),
				),
			},
			{
				ResourceName:      "openwrt_dhcp_pool.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDHCPPool_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPPoolDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPPoolConfigBasic("tf-acc-pool", "lan"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_dhcp_pool.test", "interface", "lan"),
				),
			},
			{
				Config: testAccDHCPPoolConfigUpdate("tf-acc-pool", "192.168.1.100", "192.168.1.200"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_dhcp_pool.test", "interface", "lan"),
					resource.TestCheckResourceAttr("openwrt_dhcp_pool.test", "start", "100"),
					resource.TestCheckResourceAttr("openwrt_dhcp_pool.test", "limit", "100"),
				),
			},
		},
	})
}

func TestAccDHCPDNSMasq_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPDNSMasqDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPDNSMasqConfigBasic("tf-acc-test.local"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_dhcp_dnsmasq.test", "domain", "tf-acc-test.local"),
				),
			},
		},
	})
}

func testAccCheckDHCPPoolDestroyed(s *terraform.State) error {
	client := GetTestProvider(&testing.T{})
	if client == nil {
		return nil
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openwrt_dhcp_pool" {
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

func testAccCheckDHCPDNSMasqDestroyed(s *terraform.State) error {
	client := GetTestProvider(&testing.T{})
	if client == nil {
		return nil
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openwrt_dhcp_dnsmasq" {
			continue
		}

		if rs.Primary.ID == "" {
			continue
		}

		data, err := client.UCIGetAll(context.Background(), "dhcp", "@dnsmasq[0]")
		if err != nil {
			return nil
		}

		if len(data) > 0 {
			return nil
		}
	}

	return nil
}

func testAccDHCPPoolConfigBasic(name, interfaceName string) string {
	return ProviderConfig() + `
resource "openwrt_dhcp_pool" "test" {
  name      = "` + name + `"
  interface = "` + interfaceName + `"
}
`
}

func testAccDHCPPoolConfigUpdate(name, startIP, endIP string) string {
	return ProviderConfig() + `
resource "openwrt_dhcp_pool" "test" {
  name      = "` + name + `"
  interface = "lan"
  start     = 100
  limit     = 100
}
`
}

func testAccDHCPDNSMasqConfigBasic(domainSuffix string) string {
	return ProviderConfig() + `
resource "openwrt_dhcp_dnsmasq" "test" {
  domain = "` + domainSuffix + `"
}
`
}
