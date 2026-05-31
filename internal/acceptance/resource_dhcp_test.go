package acceptance

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDHCPPool_basic(t *testing.T) {
	dhcpInterface := GetDHCPInterface()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPPoolDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPPoolConfigBasic("tf-acc-pool", dhcpInterface),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_dhcp_pool.test", "interface", dhcpInterface),
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
	dhcpInterface := GetDHCPInterface()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPPoolDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPPoolConfigBasic("tf-acc-pool", dhcpInterface),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_dhcp_pool.test", "interface", dhcpInterface),
				),
			},
			{
				Config: testAccDHCPPoolConfigUpdate("tf-acc-pool", "192.168.1.100", "192.168.1.200"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_dhcp_pool.test", "interface", dhcpInterface),
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
	dhcpInterface := GetDHCPInterface()
	return ProviderConfig() + `
resource "openwrt_dhcp_pool" "test" {
  name      = "` + name + `"
  interface = "` + dhcpInterface + `"
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

func TestAccDHCPHost_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPHostDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPHostConfigBasic("tf-acc-host", "192.168.1.150"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_dhcp_host.test", "name", "tf-acc-host"),
					resource.TestCheckResourceAttr("openwrt_dhcp_host.test", "ip", "192.168.1.150"),
				),
			},
			{
				ResourceName:      "openwrt_dhcp_host.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDHCPHost_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPHostDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPHostConfigBasic("tf-acc-host", "192.168.1.150"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_dhcp_host.test", "name", "tf-acc-host"),
					resource.TestCheckResourceAttr("openwrt_dhcp_host.test", "ip", "192.168.1.150"),
				),
			},
			{
				Config: testAccDHCPHostConfigUpdate("tf-acc-host"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_dhcp_host.test", "name", "tf-acc-host"),
					resource.TestCheckResourceAttr("openwrt_dhcp_host.test", "ip", "192.168.1.200"),
				),
			},
		},
	})
}

func testAccCheckDHCPHostDestroyed(s *terraform.State) error {
	client := GetTestProvider(&testing.T{})
	if client == nil {
		return nil
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openwrt_dhcp_host" {
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

func testAccDHCPHostConfigBasic(name, ip string) string {
	return ProviderConfig() + `
resource "openwrt_dhcp_host" "test" {
  name = "` + name + `"
  ip   = "` + ip + `"
  mac  = ["52:54:00:12:35:79"]
}
`
}

func testAccDHCPHostConfigUpdate(name string) string {
	return ProviderConfig() + `
resource "openwrt_dhcp_host" "test" {
  name      = "` + name + `"
  ip        = "192.168.1.200"
  mac       = ["52:54:00:12:35:79"]
  leasetime = "24h"
}
`
}
