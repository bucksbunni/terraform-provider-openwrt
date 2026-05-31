package acceptance

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestConfig describes the device/interface names that are safe to create, modify,
// and destroy during acceptance tests. Values are read from testconfig.yaml (in the
// package directory) when present; otherwise the built-in defaults are used.
//
// The defaults target the throwaway OpenWrt VM in testinfra/. When running against a
// real router, supply testconfig.yaml so tests never touch interfaces the router
// needs for its own connectivity — see RequireTestConfig.
type TestConfig struct {
	BridgeDevices []string       `yaml:"bridge_devices"`
	VLANConfig    VLANConfig     `yaml:"vlan_config"`
	Wireless      WirelessConfig `yaml:"wireless"`
	DHCP          DHCPConfig     `yaml:"dhcp"`
	Firewall      FirewallConfig `yaml:"firewall"`
}

type VLANConfig struct {
	BridgeDevice string   `yaml:"bridge_device"`
	TaggedPorts  []string `yaml:"tagged_ports"`
}

type WirelessConfig struct {
	Radio string `yaml:"radio"`
}

type DHCPConfig struct {
	Interface string `yaml:"interface"`
}

type FirewallConfig struct {
	Zones []string `yaml:"zones"`
}

// configPath is the optional per-environment test configuration, resolved relative
// to the package directory the tests run in.
const configPath = "testconfig.yaml"

var (
	testConfig         *TestConfig
	testConfigFromFile bool
)

func loadTestConfig() *TestConfig {
	if testConfig != nil {
		return testConfig
	}

	testConfig = &TestConfig{
		BridgeDevices: []string{"br-test"},
		VLANConfig: VLANConfig{
			BridgeDevice: "br-test",
			TaggedPorts:  []string{"eth2", "eth3"},
		},
		Wireless: WirelessConfig{
			Radio: "radio0",
		},
		DHCP: DHCPConfig{
			Interface: "lan",
		},
		Firewall: FirewallConfig{
			Zones: []string{"lan", "wan"},
		},
	}

	if _, err := os.Stat(configPath); err == nil {
		if data, readErr := os.ReadFile(configPath); readErr == nil {
			if yaml.Unmarshal(data, testConfig) == nil {
				testConfigFromFile = true
			}
		}
	}

	return testConfig
}

// RequireTestConfig skips the test unless a testconfig.yaml is present. Guard tests
// that create/modify/destroy real network devices with it, so running against a
// physical router without a vetted config can't disrupt connectivity by acting on
// the built-in default device names.
func RequireTestConfig(t *testing.T) {
	loadTestConfig()
	if !testConfigFromFile {
		t.Skip("skipping device-mutating test: no testconfig.yaml present " +
			"(copy testconfig.example.yaml to internal/acceptance/testconfig.yaml and set safe device names)")
	}
}

func GetBridgeDevice() string {
	return loadTestConfig().BridgeDevices[0]
}

func GetVLANBridgeDevice() string {
	return loadTestConfig().VLANConfig.BridgeDevice
}

func GetVLANTaggedPorts() []string {
	return loadTestConfig().VLANConfig.TaggedPorts
}

func GetWirelessRadio() string {
	return loadTestConfig().Wireless.Radio
}

func GetDHCPInterface() string {
	return loadTestConfig().DHCP.Interface
}

func GetFirewallSrcZone() string {
	zones := loadTestConfig().Firewall.Zones
	if len(zones) > 0 {
		return zones[0]
	}
	return "lan"
}

func GetFirewallDestZone() string {
	zones := loadTestConfig().Firewall.Zones
	if len(zones) > 1 {
		return zones[1]
	}
	return "wan"
}
