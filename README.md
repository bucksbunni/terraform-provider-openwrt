# Terraform Provider for OpenWrt (JSON‑RPC)

This provider manages OpenWrt devices via the LuCI JSON‑RPC API.

It wraps the following LuCI RPC libraries:

- `/cgi-bin/luci/rpc/uci` – UCI configuration
- `/cgi-bin/luci/rpc/fs` – filesystem operations
- `/cgi-bin/luci/rpc/sys` – system info and utilities
- `/cgi-bin/luci/rpc/ipkg` – package manager (opkg)

The upstream JSON‑RPC behaviour is documented in `JsonRpcHowTo.md` of the [LuCI Wiki](https://github.com/openwrt/luci/wiki).

> **WARNING: All features are highly experimental!** Full testing has not been done yet. Use in production at your own risk.

## Requirements

On the router:

```sh
opkg update
opkg install luci-mod-rpc luci-lib-ipkg luci-compat
/etc/init.d/uhttpd restart
```

Ensure LuCI is reachable at something like:

- http://192.168.1.1/cgi-bin/luci
- or https://router.example/cgi-bin/luci

## Installation

Build the provider binary:

```bash
go build ./...
```

Place `terraform-provider-openwrt` in your Terraform plugin directory or use terraform init with `source = "bucksbunni/openwrt"` once this provider becomes available.

## Provider configuration

```hcl
terraform {
  required_providers {
    openwrt = {
      source  = "bucksbunni/openwrt"
      version = "0.1.0"
    }
  }
}

provider "openwrt" {
  host     = "http://192.168.1.1"
  username = "root"
  password = "yourpassword"
  insecure = true # only if using self-signed HTTPS
}
```

`host` must point at the LuCI base URL (without `/rpc/...` appended).

## Resources

### Generic Resources

#### openwrt_uci_section

Manage a UCI section and its options:

```hcl
resource "openwrt_uci_section" "lan" {
  config = "network"
  type   = "interface"
  name   = "lan"

  options = {
    ifname  = "eth0"
    proto   = "static"
    ipaddr  = "192.168.1.1"
    netmask = "255.255.255.0"
  }
}
```

This uses `/rpc/uci`:
- `section` to create/update
- `tset` to update
- `commit` and `apply` to persist & apply changes

#### openwrt_fs_file

Manage a file via `/rpc/fs`:

```hcl
resource "openwrt_fs_file" "motd" {
  path    = "/etc/motd"
  content = "Managed by Terraform\n"
}
```

Content is base64‑encoded on the wire; the resource works with plain UTF‑8.

#### openwrt_ipkg_package

Ensure a package is installed using `/rpc/ipkg`:

```hcl
resource "openwrt_ipkg_package" "luci_mod_rpc" {
  name = "luci-mod-rpc"
}
```

### Network Resources

#### openwrt_network_interface

Manages a network interface (LAN, WAN, etc.):

```hcl
resource "openwrt_network_interface" "lan" {
  name    = "lan"
  proto   = "static"
  device  = "br-lan"
  ipaddr  = "192.168.1.1/24"
  gateway = "192.168.1.254"
  dns     = "8.8.8.8 8.8.4.4"
  metric  = 100
  auto    = "1"
}
```

#### openwrt_network_device

Manages network devices (bridges, bonding):

```hcl
resource "openwrt_network_device" "br_lan" {
  name  = "br-lan"
  type  = "bridge"
  ports = "eth0 eth1"
}

resource "openwrt_network_device" "bond0" {
  name            = "bond0"
  type            = "bonding"
  ports           = "eth1 eth2"
  policy          = "802.3ad"
  xmit_hash_policy = "layer2+3"
}
```

#### openwrt_network_bridge_vlan

Manages bridge VLAN assignments:

```hcl
resource "openwrt_network_bridge_vlan" "vlan1" {
  device = "br-lan"
  vlan   = 1
  ports  = "eth0:u* eth1:u*"
}

resource "openwrt_network_bridge_vlan" "vlan10" {
  device = "br-lan"
  vlan   = 10
  ports  = "eth2:t eth3:t"
}
```

#### openwrt_network_globals

Manages network global settings:

```hcl
resource "openwrt_network_globals" "main" {
  ula_prefix     = "fda1:b7f5:46f7::/48"
  packet_steering = true
}
```

#### openwrt_network_wireguard

Manages WireGuard VPN peers:

```hcl
resource "openwrt_network_wireguard" "peer1" {
  name                  = "wgclient1"
  description           = "Remote peer 1"
  public_key            = "ABC123...="
  endpoint_host         = "vpn.example.com"
  endpoint_port         = 51820
  persistent_keepalive  = 25
  allowed_ips           = "10.0.0.2/32 10.0.0.0/24"
}
```

### Firewall Resources

#### openwrt_firewall_zone

Manages firewall zones:

```hcl
resource "openwrt_firewall_zone" "lan" {
  name     = "lan"
  input    = "ACCEPT"
  output   = "ACCEPT"
  forward  = "REJECT"
  masq     = false
  network  = "lan"
}
```

#### openwrt_firewall_rule

Manages firewall rules:

```hcl
resource "openwrt_firewall_rule" "allow_dns" {
  name      = "Allow-DNS"
  src       = "guest"
  dest_port = "53"
  proto     = "tcpudp"
  target    = "ACCEPT"
}

resource "openwrt_firewall_rule" "allow_http" {
  name        = "Allow-HTTP"
  src         = "wan"
  dest        = "lan"
  dest_port   = "80"
  proto       = "tcp"
  target      = "ACCEPT"
  src_ip      = "192.168.1.0/24"
}
```

#### openwrt_firewall_forwarding

Manages zone-to-zone traffic forwarding:

```hcl
resource "openwrt_firewall_forwarding" "lan_to_wan" {
  src  = "lan"
  dest = "wan"
}

resource "openwrt_firewall_forwarding" "guest_to_wan" {
  src  = "guest"
  dest = "wan"
}
```

#### openwrt_firewall_defaults

Manages firewall default policies:

```hcl
resource "openwrt_firewall_defaults" "main" {
  input             = "REJECT"
  output            = "ACCEPT"
  forward           = "REJECT"
  synflood_protect  = true
  synflood_rate     = "25/s"
  drop_invalid     = true
  auto_helper      = true
}
```

### DHCP Resources

#### openwrt_dhcp_pool

Manages DHCP pools for network interfaces:

```hcl
resource "openwrt_dhcp_pool" "lan" {
  name      = "lan"
  interface = "lan"
  start     = 100
  limit     = 150
  leasetime = "12h"
  dhcpv4    = "server"
  ra        = "server"
  ra_flags  = "managed-config other-config"
}
```

#### openwrt_dhcp_dnsmasq

Manages DNS/DHCP server (dnsmasq) settings:

```hcl
resource "openwrt_dhcp_dnsmasq" "main" {
  domainneeded     = true
  rebind_protection = true
  rebind_localhost  = true
  local            = "/lan/"
  domain           = "lan"
  expand_hosts     = true
  cachesize        = 1000
  authoritative    = true
  readethers       = true
  leasefile        = "/tmp/dhcp.leases"
  resolvfile       = "/tmp/resolv.conf.d/resolv.conf.auto"
  localservice     = true
  ednspacket_max   = 1232
  server           = "8.8.8.8 8.8.4.4"
}
```

#### openwrt_dhcp_host

Manages static DHCP host reservations:

```hcl
resource "openwrt_dhcp_host" "server" {
  name      = "fileserver"
  ip        = "192.168.1.100"
  mac       = "00:11:22:33:44:55"
  leasetime = "infinite"
}
```

#### openwrt_dhcp_odhcpd

Manages DHCPv6 (odhcpd) settings:

```hcl
resource "openwrt_dhcp_odhcpd" "main" {
  maindhcp     = false
  leasefile   = "/tmp/hosts/odhcpd"
  leasetrigger = "/usr/sbin/odhcpd-update"
  loglevel    = 4
}
```

### Wireless Resources

> **Prerequisite**: Wireless support requires kernel modules and firmware packages to be installed on the router. These vary by hardware chipset (Atheros, Broadcom, MediaTek, Intel, etc.) and must be installed separately using `openwrt_ipkg_package`.

**Typical setup for Atheros hardware:**

```hcl
# Install kernel module and firmware
resource "openwrt_ipkg_package" "ath10k_dkmod" {
  name = "kmod-ath10k"
}

resource "openwrt_ipkg_package" "ath10k_fw" {
  name = "ath10k-firmware-qca988x"
}

# Restart network to load the driver
resource "openwrt_sys_rpc" "wifi_restart" {
  method      = "init.restart"
  params_json = jsonencode(["network"])
}

# Now configure wireless
resource "openwrt_wireless_device" "radio0" {
  # ...
}
```

#### openwrt_wireless_device

Manages wireless radio devices:

```hcl
resource "openwrt_wireless_device" "radio0" {
  name     = "radio0"
  type     = "mac80211"
  path     = "pci0000:00/0000:00:02.5/0000:04:00.0"
  band     = "5g"
  channel  = 44
  htmode   = "VHT80"
  country  = "DE"
  disabled = false
}
```

#### openwrt_wireless_iface

Manages wireless interfaces (SSIDs):

```hcl
resource "openwrt_wireless_iface" "main" {
  name        = "wifinet0"
  device      = "radio0"
  mode        = "ap"
  ssid        = "MyNetwork"
  encryption  = "psk2"
  key         = "secretpassword"
  network     = "lan"
  hidden      = false
  isolate     = false
}
```

### System Resources

#### openwrt_system

Manages system settings:

```hcl
resource "openwrt_system" "main" {
  hostname     = "openwrt-router"
  ttylogin     = false
  log_size     = 128
  urandom_seed = true
  zonename     = "UTC"
  log_proto    = "udp"
  conloglevel  = "8"
  cronloglevel = "7"
}
```

#### openwrt_system_ntp

Manages NTP (timeserver) configuration:

```hcl
resource "openwrt_system_ntp" "main" {
  name    = "ntp"
  enabled = true
  server  = "0.openwrt.pool.ntp.org 1.openwrt.pool.ntp.org"
}
```

#### openwrt_system_led

Manages system LED configuration:

```hcl
resource "openwrt_system_led" "wan_led" {
  name     = "WAN LED"
  sysfs   = "apu:green:3"
  trigger = "netdev"
  mode    = "link tx rx"
  dev     = "eth0"
}

resource "openwrt_system_led" "diag_led" {
  name     = "DIAG"
  sysfs   = "apu:green:1"
  trigger = "none"
  default = true
}
```

### SSH Resources

#### openwrt_dropbear

Manages Dropbear SSH server settings:

```hcl
resource "openwrt_dropbear" "main" {
  name             = "main"
  password_auth    = false
  port             = 22
  root_password_auth = true
  root_login       = true
}
```

### Web Server Resources

#### openwrt_uhttpd

Manages uHTTPd web server settings:

```hcl
resource "openwrt_uhttpd" "main" {
  name              = "main"
  listen_http       = "0.0.0.0:80 [::]:80"
  listen_https     = "0.0.0.0:443 [::]:443"
  redirect_https    = true
  home              = "/www"
  rfc1918_filter   = true
  max_requests     = 3
  max_connections  = 100
  cert             = "/etc/uhttpd.crt"
  key              = "/etc/uhttpd.key"
}
```

#### openwrt_uhttpd_cert

Manages uHTTPd certificate generation settings:

```hcl
resource "openwrt_uhttpd_cert" "defaults" {
  name       = "defaults"
  days       = 397
  key_type   = "ec"
  ec_curve   = "P-256"
  country    = "US"
  state      = "California"
  location   = "San Francisco"
  commonname = "router.local"
}
```

### RPC Daemon Resources

#### openwrt_rpcd

Manages RPC daemon (ubus) settings:

```hcl
resource "openwrt_rpcd" "main" {
  socket  = "/var/run/ubus/ubus.sock"
  timeout = 30
}
```

### Data Sources

#### openwrt_sys_rpc

Low‑level `/rpc/sys` access for any sys API method:

```hcl
data "openwrt_sys_rpc" "example" {
  method      = "net.routes"
  params_json = "[]"
}

output "result" {
  value = jsondecode(data.openwrt_sys_rpc.example.result_json)
}
```

#### openwrt_sys_hostname

Retrieves the system hostname:

```hcl
data "openwrt_sys_hostname" "main" {}

output "hostname" {
  value = data.openwrt_sys_hostname.main.hostname
}
```

#### openwrt_sys_uptime

Retrieves the system uptime:

```hcl
data "openwrt_sys_uptime" "main" {}

output "uptime_seconds" {
  value = data.openwrt_sys_uptime.main.uptime
}
```

#### openwrt_sys_init

Retrieves the list of init scripts and their status:

```hcl
data "openwrt_sys_init" "main" {}

output "enabled_services" {
  value = [for s in data.openwrt_sys_init.main.scripts : s.name if s.enabled]
}
```

#### openwrt_sys_net_devices

Retrieves network device information:

```hcl
data "openwrt_sys_net_devices" "main" {}

output "interfaces" {
  value = data.openwrt_sys_net_devices.main.devices
}
```

#### openwrt_sys_net_routes

Retrieves the IPv4 routing table:

```hcl
data "openwrt_sys_net_routes" "main" {}

output "routes" {
  value = data.openwrt_sys_net_routes.main.routes
}
```

#### openwrt_sys_net_routes6

Retrieves the IPv6 routing table:

```hcl
data "openwrt_sys_net_routes6" "main" {}

output "ipv6_routes" {
  value = data.openwrt_sys_net_routes6.main.routes
}
```

#### openwrt_sys_net_arptable

Retrieves the ARP table:

```hcl
data "openwrt_sys_net_arptable" "main" {}

output "arp_entries" {
  value = data.openwrt_sys_net_arptable.main.entries
}
```

#### openwrt_sys_net_conntrack

Retrieves the connection tracking table:

```hcl
data "openwrt_sys_net_conntrack" "main" {}

output "connections" {
  value = data.openwrt_sys_net_conntrack.main.entries
}
```

#### openwrt_sys_process_list

Retrieves the list of running processes:

```hcl
data "openwrt_sys_process_list" "main" {}

output "processes" {
  value = data.openwrt_sys_process_list.main.processes
}
```

#### openwrt_sys_wireless

Retrieves wireless interface information:

```hcl
data "openwrt_sys_wireless" "wlan0" {
  ifname = "wlan0"
}

output "signal_strength" {
  value = data.openwrt_sys_wireless.wlan0.signal
}
```

## Examples

Example Terraform configurations are available in the [`examples/`](./examples/) directory:

- [`basic`](./examples/basic/) - Data source usage (hostname, uptime, network devices)
- [`network-interfaces`](./examples/network-interfaces/) - Network bridges, interfaces, and VLANs
- [`firewall-setup`](./examples/firewall-setup/) - Multi-zone firewall with rules and forwarding
- [`dhcp-dns`](./examples/dhcp-dns/) - DHCP pools, dnsmasq, and static reservations
- [`wireless`](./examples/wireless/) - Dual-band WiFi with multiple SSIDs
- [`system`](./examples/system/) - System settings, NTP, LEDs, and SSH
- [`complete-router`](./examples/complete-router/) - Full router configuration

See [`examples/README.md`](./examples/README.md) for usage instructions.

## Limitations and TODOs

- LuCI path is fixed to /cgi-bin/luci; making it configurable is a future enhancement.
- TLS options are limited to insecure; supporting CA bundles and client certs would be useful for production.
- Unit tests for typed resources are not yet implemented.
- Acceptance tests require a live OpenWrt device.
- Import support could be extended for additional resources.

## License

This project has been developed under the [GNU GPLv3](./LICENSE) license.
