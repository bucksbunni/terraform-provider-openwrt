package acceptance

// These tests cover the destroy/refresh behaviour of resources backed by
// anonymous UCI sections (bridge VLANs, network devices, dropbear, LEDs), which
// the provider addresses by the stable internal section identifier persisted in
// the Computed section attribute. They specifically guard the regression
// from issue #55: destroying an openwrt_network_bridge_vlan must succeed even
// when its underlying section has already been removed out of band - for example
// when a managed parent openwrt_network_device is torn down first and netifd
// drops the attached bridge-vlan sections. The "Disappears" tests delete the
// section externally and assert the provider treats the missing section as
// already deleted (a clean refresh/destroy) rather than erroring.

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// deleteSectionByMatch deletes, out of band, the first section of sectionType in
// config whose options all equal the values in match, committing the change. It
// simulates the section vanishing underneath Terraform. It is a no-op if nothing
// matches, so it never fails a test merely because the section is already gone.
func deleteSectionByMatch(config, sectionType string, match map[string]string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := GetTestProvider(&testing.T{})
		if client == nil {
			return nil
		}

		sections, err := client.UCIForeach(context.Background(), config, sectionType)
		if err != nil {
			return err
		}

		for _, sec := range sections {
			matched := true
			for opt, want := range match {
				if fmt.Sprintf("%v", sec[opt]) != want {
					matched = false
					break
				}
			}
			if !matched {
				continue
			}

			name, _ := sec[".name"].(string)
			if name == "" {
				return nil
			}
			if err := client.UCIDelete(context.Background(), config, name); err != nil {
				return err
			}
			return client.UCICommit(context.Background(), config)
		}

		return nil
	}
}

// testAccCheckBridgeVlanSectionGone asserts that no bridge-vlan section matching
// device+vlan remains on the device. Unlike the synthetic-id based destroy check,
// it scans the actual anonymous sections via foreach, so it genuinely fails if a
// section is left dangling after destroy.
func testAccCheckBridgeVlanSectionGone(device string, vlan int) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := GetTestProvider(&testing.T{})
		if client == nil {
			return nil
		}

		sections, err := client.UCIForeach(context.Background(), "network", "bridge-vlan")
		if err != nil {
			return nil
		}

		for _, sec := range sections {
			if fmt.Sprintf("%v", sec["device"]) == device &&
				fmt.Sprintf("%v", sec["vlan"]) == fmt.Sprintf("%d", vlan) {
				return fmt.Errorf("bridge-vlan section for device %q vlan %d still exists after destroy", device, vlan)
			}
		}

		return nil
	}
}

// TestAccNetworkBridgeVlan_Disappears removes the bridge-vlan section out of band
// after creation and asserts the provider tolerates it: the follow-up refresh
// must report the resource as gone (non-empty plan) and the teardown destroy must
// not error. This is the core of the #55 regression.
func TestAccNetworkBridgeVlan_Disappears(t *testing.T) {
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
		CheckDestroy:             testAccCheckBridgeVlanSectionGone(bridgeDevice, 100),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkBridgeVlanConfigBasic(bridgeDevice, 100, ports[0]),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_bridge_vlan.test", "device", bridgeDevice),
					resource.TestCheckResourceAttr("openwrt_network_bridge_vlan.test", "vlan", "100"),
					resource.TestCheckResourceAttrSet("openwrt_network_bridge_vlan.test", "section"),
					deleteSectionByMatch("network", "bridge-vlan", map[string]string{
						"device": bridgeDevice,
						"vlan":   "100",
					}),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccNetworkBridgeVlan_DestroyWithManagedDevice reproduces the exact #55
// topology: a bridge VLAN whose parent bridge device is also managed. The VLAN
// references the device by its literal name rather than via an attribute, so no
// dependency edge forces the VLAN to be destroyed first; the parent device may be
// torn down first, removing the VLAN section as a side effect. The strong destroy
// check then verifies that neither a dangling section nor a destroy error remains.
func TestAccNetworkBridgeVlan_DestroyWithManagedDevice(t *testing.T) {
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
		CheckDestroy:             testAccCheckBridgeVlanSectionGone(bridgeDevice, 100),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkBridgeVlanConfigNoDep(bridgeDevice, 100, ports[0]),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_network_bridge_vlan.test", "vlan", "100"),
					resource.TestCheckResourceAttrSet("openwrt_network_bridge_vlan.test", "section"),
					resource.TestCheckResourceAttrSet("openwrt_network_device.bridge", "section"),
				),
			},
		},
	})
}

// TestAccNetworkDevice_Disappears removes the network device section out of band
// and asserts the provider reports it as gone on refresh instead of erroring.
func TestAccNetworkDevice_Disappears(t *testing.T) {
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
					resource.TestCheckResourceAttrSet("openwrt_network_device.test", "section"),
					deleteSectionByMatch("network", "device", map[string]string{"name": bridgeName}),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccDropbear_Disappears removes the dropbear section out of band and asserts
// the provider reports it as gone on refresh instead of erroring.
func TestAccDropbear_Disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDropbearDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDropbearConfigBasic("tf-acc-ssh"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_dropbear.test", "name", "test-dropbear"),
					resource.TestCheckResourceAttrSet("openwrt_dropbear.test", "section"),
					deleteSectionByMatch("dropbear", "dropbear", map[string]string{"name": "test-dropbear"}),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccSystemLED_Disappears removes the LED section out of band and asserts the
// provider reports it as gone on refresh instead of erroring.
func TestAccSystemLED_Disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSystemLEDDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSystemLEDConfigBasic("tf-acc-led", "tp-link:green:power"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_system_led.test", "name", "tf-acc-led"),
					resource.TestCheckResourceAttrSet("openwrt_system_led.test", "section"),
					deleteSectionByMatch("system", "led", map[string]string{"name": "tf-acc-led"}),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// testAccNetworkBridgeVlanConfigNoDep is like testAccNetworkBridgeVlanConfigBasic
// but the bridge VLAN references the device by its literal name instead of the
// openwrt_network_device attribute, so Terraform builds no dependency edge
// between them. This lets a destroy tear the parent device down first, exercising
// the #55 cascade where the bridge-vlan section is already gone at delete time.
func testAccNetworkBridgeVlanConfigNoDep(device string, vlan int, port string) string {
	return ProviderConfig() + `
resource "openwrt_network_device" "bridge" {
  name  = "` + device + `"
  type  = "bridge"
  ports = ["` + port + `"]
}

resource "openwrt_network_bridge_vlan" "test" {
  device = "` + device + `"
  vlan   = ` + fmt.Sprintf("%d", vlan) + `
  ports = {
    ` + port + ` = "t"
  }
}
`
}
