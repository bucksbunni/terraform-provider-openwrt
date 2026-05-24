package acceptance

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccNetworkInterface_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckNetworkInterfaceDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkInterfaceConfigBasic("tf-acc-lan", "static"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_interface.test", "name", "tf-acc-lan"),
					resource.TestCheckResourceAttr("openwrt_network_interface.test", "proto", "static"),
				),
			},
			{
				ResourceName:      "openwrt_network_interface.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkInterface_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckNetworkInterfaceDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkInterfaceConfigBasic("tf-acc-lan", "static"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_interface.test", "name", "tf-acc-lan"),
					resource.TestCheckResourceAttr("openwrt_network_interface.test", "proto", "static"),
				),
			},
			{
				Config: testAccNetworkInterfaceConfigUpdate("tf-acc-lan"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_interface.test", "name", "tf-acc-lan"),
					resource.TestCheckResourceAttr("openwrt_network_interface.test", "ipaddr", "192.168.100.1"),
					resource.TestCheckResourceAttr("openwrt_network_interface.test", "netmask", "255.255.255.0"),
				),
			},
		},
	})
}

func TestAccNetworkDevice_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckNetworkDeviceDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkDeviceConfigBasic("tf-acc-br", "br-lan"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_device.test", "name", "tf-acc-br"),
					resource.TestCheckResourceAttr("openwrt_network_device.test", "type", "bridge"),
				),
			},
			{
				ResourceName:      "openwrt_network_device.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckNetworkInterfaceDestroyed(s *terraform.State) error {
	client := GetTestProvider(&testing.T{})
	if client == nil {
		return nil
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openwrt_network_interface" {
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

func testAccCheckNetworkDeviceDestroyed(s *terraform.State) error {
	client := GetTestProvider(&testing.T{})
	if client == nil {
		return nil
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openwrt_network_device" {
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

func testAccNetworkInterfaceConfigBasic(name, proto string) string {
	return ProviderConfig() + `
resource "openwrt_network_interface" "test" {
  name   = "` + name + `"
  proto  = "` + proto + `"
}
`
}

func testAccNetworkInterfaceConfigUpdate(name string) string {
	return ProviderConfig() + `
resource "openwrt_network_interface" "test" {
  name    = "` + name + `"
  proto   = "static"
  ipaddr  = "192.168.100.1"
  netmask = "255.255.255.0"
}
`
}

func testAccNetworkDeviceConfigBasic(name, device string) string {
	return ProviderConfig() + `
resource "openwrt_network_device" "test" {
  name   = "` + name + `"
  type   = "bridge"
}
`
}
