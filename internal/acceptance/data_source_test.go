package acceptance

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceSysHostname_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheckWithConnectivity(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSysHostnameConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.openwrt_sys_hostname.test", "hostname"),
				),
			},
		},
	})
}

func TestAccDataSourceSysUptime_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheckWithConnectivity(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSysUptimeConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.openwrt_sys_uptime.test", "uptime"),
				),
			},
		},
	})
}

func TestAccDataSourceSysNetDevices_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheckWithConnectivity(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSysNetDevicesConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.openwrt_sys_net_devices.test", "devices.#"),
				),
			},
		},
	})
}

func TestAccDataSourceSysNetRoutes_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheckWithConnectivity(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSysNetRoutesConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.openwrt_sys_net_routes.test", "routes.#"),
				),
			},
		},
	})
}

func TestAccDataSourceSysNetRoutes6_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheckWithConnectivity(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSysNetRoutes6Config(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.openwrt_sys_net_routes6.test", "routes.#"),
				),
			},
		},
	})
}

func TestAccDataSourceSysNetArpTable_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheckWithConnectivity(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSysNetArpTableConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.openwrt_sys_net_arptable.test", "entries.#"),
				),
			},
		},
	})
}

func TestAccDataSourceSysNetConntrack_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheckWithConnectivity(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSysNetConntrackConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.openwrt_sys_net_conntrack.test", "entries.#"),
				),
			},
		},
	})
}

func TestAccDataSourceSysProcessList_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheckWithConnectivity(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSysProcessListConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.openwrt_sys_process_list.test", "processes.#"),
				),
			},
		},
	})
}

func TestAccDataSourceSysWireless_basic(t *testing.T) {
	// This data source queries iwinfo for a specific wireless interface, which
	// requires wireless packages (iwinfo) and a radio. Provide a real interface
	// name via OPENWRT_WIRELESS_IFNAME to run it; otherwise skip.
	ifname := os.Getenv("OPENWRT_WIRELESS_IFNAME")
	if ifname == "" {
		t.Skip("skipping wireless data source test: set OPENWRT_WIRELESS_IFNAME to a wireless interface (e.g. wlan0)")
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheckWithConnectivity(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSysWirelessConfig(ifname),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.openwrt_sys_wireless.test", "ifname", ifname),
				),
			},
		},
	})
}

func TestAccDataSourceSysRPC_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheckWithConnectivity(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSysRPCConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.openwrt_sys_rpc.test", "result_json"),
				),
			},
		},
	})
}

func testAccDataSourceSysHostnameConfig() string {
	return ProviderConfig() + `
data "openwrt_sys_hostname" "test" {}
`
}

func testAccDataSourceSysUptimeConfig() string {
	return ProviderConfig() + `
data "openwrt_sys_uptime" "test" {}
`
}

func testAccDataSourceSysNetDevicesConfig() string {
	return ProviderConfig() + `
data "openwrt_sys_net_devices" "test" {}
`
}

func testAccDataSourceSysNetRoutesConfig() string {
	return ProviderConfig() + `
data "openwrt_sys_net_routes" "test" {}
`
}

func testAccDataSourceSysNetRoutes6Config() string {
	return ProviderConfig() + `
data "openwrt_sys_net_routes6" "test" {}
`
}

func testAccDataSourceSysNetArpTableConfig() string {
	return ProviderConfig() + `
data "openwrt_sys_net_arptable" "test" {}
`
}

func testAccDataSourceSysNetConntrackConfig() string {
	return ProviderConfig() + `
data "openwrt_sys_net_conntrack" "test" {}
`
}

func testAccDataSourceSysProcessListConfig() string {
	return ProviderConfig() + `
data "openwrt_sys_process_list" "test" {}
`
}

func testAccDataSourceSysWirelessConfig(ifname string) string {
	return ProviderConfig() + `
data "openwrt_sys_wireless" "test" {
  ifname = "` + ifname + `"
}
`
}

func testAccDataSourceSysRPCConfig() string {
	return ProviderConfig() + `
data "openwrt_sys_rpc" "test" {
  method = "net.devices"
}
`
}
