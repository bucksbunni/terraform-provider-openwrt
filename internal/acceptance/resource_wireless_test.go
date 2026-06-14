package acceptance

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccWirelessDevice_basic(t *testing.T) {
	RequireTestConfig(t)
	RequireWireless(t)
	radio := GetWirelessRadio()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckWirelessDeviceDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccWirelessDeviceConfigBasic(radio),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_wireless_device.test", "name", radio),
					resource.TestCheckResourceAttr("openwrt_wireless_device.test", "type", "mac80211"),
				),
			},
			{
				ResourceName:      "openwrt_wireless_device.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWirelessDevice_Update(t *testing.T) {
	RequireTestConfig(t)
	RequireWireless(t)
	radio := GetWirelessRadio()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckWirelessDeviceDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccWirelessDeviceConfigBasic(radio),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_wireless_device.test", "name", radio),
				),
			},
			{
				Config: testAccWirelessDeviceConfigUpdate(radio),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_wireless_device.test", "name", radio),
					resource.TestCheckResourceAttr("openwrt_wireless_device.test", "disabled", "false"),
				),
			},
		},
	})
}

func TestAccWirelessInterface_basic(t *testing.T) {
	RequireTestConfig(t)
	RequireWireless(t)
	radio := GetWirelessRadio()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckWirelessInterfaceDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccWirelessInterfaceConfigBasic("test-wifi", radio),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_wireless_iface.test", "device", radio),
					resource.TestCheckResourceAttr("openwrt_wireless_iface.test", "mode", "ap"),
				),
			},
			{
				ResourceName:      "openwrt_wireless_iface.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWirelessInterface_Update(t *testing.T) {
	RequireTestConfig(t)
	RequireWireless(t)
	radio := GetWirelessRadio()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckWirelessInterfaceDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccWirelessInterfaceConfigBasic("test-wifi", radio),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_wireless_iface.test", "device", radio),
				),
			},
			{
				Config: testAccWirelessInterfaceConfigUpdate("test-wifi", radio),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_wireless_iface.test", "device", radio),
					resource.TestCheckResourceAttr("openwrt_wireless_iface.test", "ssid", "TestNetwork"),
				),
			},
		},
	})
}

func testAccCheckWirelessDeviceDestroyed(s *terraform.State) error {
	client := GetTestProvider(&testing.T{})
	if client == nil {
		return nil
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openwrt_wireless_device" {
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
			return fmt.Errorf("%s %q still exists after destroy", rs.Type, parts[1])
		}
	}

	return nil
}

func testAccCheckWirelessInterfaceDestroyed(s *terraform.State) error {
	client := GetTestProvider(&testing.T{})
	if client == nil {
		return nil
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openwrt_wireless_iface" {
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
			return fmt.Errorf("%s %q still exists after destroy", rs.Type, parts[1])
		}
	}

	return nil
}

func testAccWirelessDeviceConfigBasic(name string) string {
	return ProviderConfig() + `
resource "openwrt_wireless_device" "test" {
  name = "` + name + `"
  type = "mac80211"
}
`
}

func testAccWirelessDeviceConfigUpdate(name string) string {
	return ProviderConfig() + `
resource "openwrt_wireless_device" "test" {
  name     = "` + name + `"
  type     = "mac80211"
  disabled = false
  channel  = 6
}
`
}

func testAccWirelessInterfaceConfigBasic(name, device string) string {
	return ProviderConfig() + `
resource "openwrt_wireless_iface" "test" {
  device = "` + device + `"
  mode   = "ap"
  ssid   = "TestAccNetwork"
}
`
}

func testAccWirelessInterfaceConfigUpdate(name, device string) string {
	return ProviderConfig() + `
resource "openwrt_wireless_iface" "test" {
  device     = "` + device + `"
  mode       = "ap"
  ssid       = "TestNetwork"
  encryption = "psk2"
  key        = "testpassword123"
}
`
}
