package acceptance

import (
	"context"
	"fmt"
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
				Config: testAccNetworkInterfaceConfigBasic("tf_acc_lan", "static"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_interface.test", "name", "tf_acc_lan"),
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
				Config: testAccNetworkInterfaceConfigBasic("tf_acc_lan", "static"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_interface.test", "name", "tf_acc_lan"),
					resource.TestCheckResourceAttr("openwrt_network_interface.test", "proto", "static"),
				),
			},
			{
				Config: testAccNetworkInterfaceConfigUpdate("tf_acc_lan"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_interface.test", "name", "tf_acc_lan"),
					resource.TestCheckResourceAttr("openwrt_network_interface.test", "ipaddr.0", "192.168.100.1/24"),
				),
			},
		},
	})
}

func TestAccNetworkDevice_basic(t *testing.T) {
	RequireTestConfig(t)
	bridgeName := GetBridgeDevice()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckNetworkDeviceDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkDeviceConfigBasic(bridgeName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_device.test", "name", bridgeName),
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
			return fmt.Errorf("%s %q still exists after destroy", rs.Type, parts[1])
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
			return fmt.Errorf("%s %q still exists after destroy", rs.Type, parts[1])
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
  ipaddr  = ["192.168.100.1/24"]
}
`
}

func TestAccNetworkInterface_DualStack(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckNetworkInterfaceDestroyed,
		Steps: []resource.TestStep{
			{
				// Create with both IPv4 and IPv6
				Config: testAccNetworkInterfaceConfigDualStack("tf_acc_dual"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_interface.test", "name", "tf_acc_dual"),
					resource.TestCheckResourceAttr("openwrt_network_interface.test", "ipaddr.0", "192.168.101.1/24"),
					resource.TestCheckResourceAttr("openwrt_network_interface.test", "ip6addr.0", "fd00:acc::1/64"),
				),
			},
			{
				// Update: remove ip6addr — verify it is cleared from UCI (not left stale)
				Config: testAccNetworkInterfaceConfigIPv4Only("tf_acc_dual"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_interface.test", "ipaddr.0", "192.168.101.1/24"),
					resource.TestCheckNoResourceAttr("openwrt_network_interface.test", "ip6addr.#"),
				),
			},
		},
	})
}

func testAccNetworkInterfaceConfigDualStack(name string) string {
	return ProviderConfig() + `
resource "openwrt_network_interface" "test" {
  name    = "` + name + `"
  proto   = "static"
  ipaddr  = ["192.168.101.1/24"]
  ip6addr = ["fd00:acc::1/64"]
}
`
}

func testAccNetworkInterfaceConfigIPv4Only(name string) string {
	return ProviderConfig() + `
resource "openwrt_network_interface" "test" {
  name   = "` + name + `"
  proto  = "static"
  ipaddr = ["192.168.101.1/24"]
}
`
}

func TestAccNetworkInterface_IP6Gateway(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckNetworkInterfaceDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkInterfaceConfigIP6Gateway("tf_acc_ip6gw"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_interface.test", "ip6gateway", "fd00:acc::1"),
					// The Terraform attribute is "ip6gateway", but netifd's static proto
					// handler only reads the UCI option "ip6gw" (proto_ip_attributes in
					// netifd's proto.c). Assert the wire-level key directly so a
					// regression back to the wrong key is caught even though Read()
					// would otherwise round-trip a wrong key back to itself.
					testAccCheckNetworkInterfaceUCIOption("openwrt_network_interface.test", "ip6gw", "fd00:acc::1"),
				),
			},
		},
	})
}

func testAccNetworkInterfaceConfigIP6Gateway(name string) string {
	return ProviderConfig() + `
resource "openwrt_network_interface" "test" {
  name       = "` + name + `"
  proto      = "static"
  ip6addr    = ["fd00:acc::2/64"]
  ip6gateway = "fd00:acc::1"
}
`
}

func testAccCheckNetworkInterfaceUCIOption(resourceName, option, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		parts := splitImportID(rs.Primary.ID)
		if len(parts) != 2 {
			return fmt.Errorf("unexpected resource ID format: %s", rs.Primary.ID)
		}

		client := GetTestProvider(&testing.T{})
		if client == nil {
			return nil
		}

		data, err := client.UCIGetAll(context.Background(), parts[0], parts[1])
		if err != nil {
			return err
		}

		got, _ := data[option].(string)
		if got != want {
			return fmt.Errorf("expected UCI option %q to be %q, got %q (raw: %#v)", option, want, got, data[option])
		}

		return nil
	}
}

func TestAccNetworkBridgeVlan_basic(t *testing.T) {
	RequireTestConfig(t)
	bridgeDevice := GetVLANBridgeDevice()
	ports := GetVLANTaggedPorts()
	if len(ports) == 0 {
		t.Skip("No tagged ports configured for VLAN tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckNetworkBridgeVlanDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkBridgeVlanConfigBasic(bridgeDevice, 100, ports[0]),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_bridge_vlan.test", "device", bridgeDevice),
					resource.TestCheckResourceAttr("openwrt_network_bridge_vlan.test", "vlan", "100"),
				),
			},
			{
				ResourceName:      "openwrt_network_bridge_vlan.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkBridgeVlan_Update(t *testing.T) {
	RequireTestConfig(t)
	bridgeDevice := GetVLANBridgeDevice()
	ports := GetVLANTaggedPorts()
	if len(ports) == 0 {
		t.Skip("No tagged ports configured for VLAN tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckNetworkBridgeVlanDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkBridgeVlanConfigBasic(bridgeDevice, 100, ports[0]),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_bridge_vlan.test", "device", bridgeDevice),
					resource.TestCheckResourceAttr("openwrt_network_bridge_vlan.test", "vlan", "100"),
				),
			},
			{
				Config: testAccNetworkBridgeVlanConfigUpdate(bridgeDevice, 100, ports[0]),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_bridge_vlan.test", "device", bridgeDevice),
					resource.TestCheckResourceAttr("openwrt_network_bridge_vlan.test", "vlan", "100"),
				),
			},
		},
	})
}

func TestAccNetworkWireguard_basic(t *testing.T) {
	RequireWireguard(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckNetworkWireguardDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkWireguardConfigBasic("wg_test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_wireguard.test", "name", "wg_test"),
					resource.TestCheckResourceAttr("openwrt_network_wireguard.test", "public_key", "test-public-key-12345"),
				),
			},
			{
				ResourceName:      "openwrt_network_wireguard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkWireguard_Update(t *testing.T) {
	RequireWireguard(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckNetworkWireguardDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkWireguardConfigBasic("wg_test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_wireguard.test", "name", "wg_test"),
				),
			},
			{
				Config: testAccNetworkWireguardConfigUpdate("wg_test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_wireguard.test", "name", "wg_test"),
					resource.TestCheckResourceAttr("openwrt_network_wireguard.test", "endpoint_host", "vpn.example.com"),
					resource.TestCheckResourceAttr("openwrt_network_wireguard.test", "endpoint_port", "51820"),
				),
			},
		},
	})
}

func testAccCheckNetworkBridgeVlanDestroyed(s *terraform.State) error {
	client := GetTestProvider(&testing.T{})
	if client == nil {
		return nil
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openwrt_network_bridge_vlan" {
			continue
		}

		if rs.Primary.ID == "" {
			continue
		}

		parts := splitImportID(rs.Primary.ID)
		if len(parts) != 2 {
			continue
		}

		data, err := client.UCIGetAll(context.Background(), "network", parts[1])
		if err != nil {
			return nil
		}

		if len(data) > 0 {
			return fmt.Errorf("%s %q still exists after destroy", rs.Type, parts[1])
		}
	}

	return nil
}

func testAccCheckNetworkWireguardDestroyed(s *terraform.State) error {
	client := GetTestProvider(&testing.T{})
	if client == nil {
		return nil
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openwrt_network_wireguard" {
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

func testAccNetworkBridgeVlanConfigBasic(device string, vlan int, port string) string {
	return ProviderConfig() + `
resource "openwrt_network_device" "bridge" {
  name  = "` + device + `"
  type  = "bridge"
  ports = ["` + port + `"]
}

resource "openwrt_network_bridge_vlan" "test" {
  device = openwrt_network_device.bridge.name
  vlan   = ` + fmt.Sprintf("%d", vlan) + `
  ports = {
    ` + port + ` = "t"
  }
}
`
}

func testAccNetworkBridgeVlanConfigUpdate(device string, vlan int, port string) string {
	return ProviderConfig() + `
resource "openwrt_network_device" "bridge" {
  name  = "` + device + `"
  type  = "bridge"
  ports = ["` + port + `"]
}

resource "openwrt_network_bridge_vlan" "test" {
  device = openwrt_network_device.bridge.name
  vlan   = ` + fmt.Sprintf("%d", vlan) + `
  ports = {
    ` + port + ` = "t"
  }
}
`
}

func testAccNetworkWireguardConfigBasic(name string) string {
	return ProviderConfig() + `
resource "openwrt_network_wireguard" "test" {
  name       = "` + name + `"
  public_key = "test-public-key-12345"
  allowed_ips = ["10.0.0.2/32"]
}
`
}

func testAccNetworkWireguardConfigUpdate(name string) string {
	return ProviderConfig() + `
resource "openwrt_network_wireguard" "test" {
  name           = "` + name + `"
  public_key     = "test-public-key-12345"
  endpoint_host  = "vpn.example.com"
  endpoint_port  = 51820
  allowed_ips    = ["10.0.0.2/32", "192.168.200.0/24"]
}
`
}

func testAccNetworkDeviceConfigBasic(name string) string {
	return ProviderConfig() + `
resource "openwrt_network_device" "test" {
  name   = "` + name + `"
  type   = "bridge"
}
`
}
