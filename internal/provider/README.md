# Provider Package

This package contains the core implementation of the Terraform provider for OpenWrt.

## Overview

The provider communicates with OpenWrt devices through the LuCI JSON-RPC API to manage system configuration using the UCI (Unified Configuration Interface) system.

## Main Components

### Provider (`provider.go`)

The main provider implementation that:

- Registers all resources and data sources
- Handles provider configuration (host, username, password)
- Creates and manages the JSON-RPC client

### JSON-RPC Client (`client.go`)

The `JsonRpcClient` provides methods for:

- **UCI Operations**: Reading, writing, and managing UCI configuration
- **Filesystem Operations**: Reading, writing, and managing files on the device
- **Package Management**: Installing and removing packages via opkg
- **System Calls**: Executing LuCI system RPC calls

### Resources

The provider implements the following Terraform resources:

#### Generic Resources

- `openwrt_uci_section` - Generic UCI section management
- `openwrt_fs_file` - Filesystem file management
- `openwrt_ipkg_package` - Package installation/removal

#### Firewall Resources

- `openwrt_firewall_zone` - Firewall zone configuration
- `openwrt_firewall_rule` - Firewall rule configuration
- `openwrt_firewall_forwarding` - Zone-to-zone forwarding
- `openwrt_firewall_defaults` - Default firewall policies

#### Network Resources

- `openwrt_network_interface` - Network interface configuration
- `openwrt_network_device` - Network device (bridge/bond) configuration
- `openwrt_network_bridge_vlan` - Bridge VLAN assignment
- `openwrt_network_globals` - Global network settings
- `openwrt_network_wireguard` - WireGuard peer configuration

#### DHCP Resources

- `openwrt_dhcp_pool` - DHCP pool configuration
- `openwrt_dhcp_dnsmasq` - DNS/DHCP (dnsmasq) settings
- `openwrt_dhcp_odhcpd` - DHCPv6 (odhcpd) settings
- `openwrt_dhcp_host` - Static DHCP host reservations

#### System Resources

- `openwrt_system` - System settings
- `openwrt_system_ntp` - NTP/timeserver configuration
- `openwrt_system_led` - LED configuration

#### Other Resources

- `openwrt_dropbear` - Dropbear SSH server
- `openwrt_wireless_device` - Wireless radio device
- `openwrt_wireless_interface` - Wireless interface (WiFi SSID)
- `openwrt_uhttpd` - uHTTPd web server
- `openwrt_uhttpd_cert` - HTTPS certificate configuration
- `openwrt_rpcd` - RPC daemon settings

### Data Sources

The provider implements the following data sources:

- `openwrt_sys_rpc` - Low-level RPC access
- `openwrt_sys_hostname` - System hostname
- `openwrt_sys_uptime` - System uptime
- `openwrt_sys_init` - Init scripts list
- `openwrt_sys_net_devices` - Network devices information
- `openwrt_sys_net_routes` - IPv4 routing table
- `openwrt_sys_net_routes6` - IPv6 routing table
- `openwrt_sys_net_arptable` - ARP table
- `openwrt_sys_net_conntrack` - Connection tracking table
- `openwrt_sys_process_list` - Running processes
- `openwrt_sys_wireless` - Wireless interface information

## Subpackages

- [`converters`](converters/) - Data model converters for UCI types
- [`importid`](importid/) - Import ID parsing utilities

## Testing

Unit tests are co-located with implementation files (e.g., `resource_firewall_zone.go` and `resource_firewall_zone_test.go`).

Run tests with:
```bash
go test ./internal/provider/...
```

## Usage Example

```hcl
provider "openwrt" {
  host     = "http://192.168.1.1"
  username = "root"
  password = "secret"
}

resource "openwrt_firewall_zone" "guest" {
  name   = "guest"
  input  = "ACCEPT"
  output = "ACCEPT"
  masq   = true
  network = "guest"
}
```
