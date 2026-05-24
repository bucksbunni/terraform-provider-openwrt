---
layout: ""
page_title: "DHCP Resources"
description: |-
  DHCP resource types for the OpenWrt provider.
---

# DHCP Resources

## openwrt_dhcp_pool

Manages DHCP pools for network interfaces.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `name` | String | Yes | Pool name (same as interface) |
| `interface` | String | Yes | Network interface |
| `start` | Int64 | No | DHCP range start |
| `limit` | Int64 | No | Number of addresses |
| `leasetime` | String | No | Lease time (e.g., '12h') |
| `dhcpv4` | String | No | DHCPv4 mode: 'server', 'relay', 'none' |
| `dhcpv6` | String | No | DHCPv6 mode: 'server', 'relay', 'none' |
| `ra` | String | No | Router advertisement: 'server', 'relay', 'none' |
| `ra_flags` | String | No | RA flags |

### Example

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

---

## openwrt_dhcp_dnsmasq

Manages DNS/DHCP server (dnsmasq) settings.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `domainneeded` | Bool | No | Require domain in queries |
| `rebind_protection` | Bool | No | Enable DNS rebind protection |
| `rebind_localhost` | Bool | No | Allow localhost in rebind |
| `local` | String | No | Local domain suffix |
| `domain` | String | No | Domain name |
| `expand_hosts` | Bool | No | Expand hosts in DNS |
| `cachesize` | Int64 | No | DNS cache size |
| `authoritative` | Bool | No | Authoritative DHCP server |
| `readethers` | Bool | No | Read /etc/ethers |
| `leasefile` | String | No | DHCP lease file |
| `resolvfile` | String | No | DNS resolvers file |
| `localservice` | Bool | No | Only local subnets |
| `ednspacket_max` | Int64 | No | EDNS packet max size |
| `server` | String | No | Upstream DNS servers |
| `noresolv` | Bool | No | Don't use resolvers |
| `strictorder` | Bool | No | Query in order |

### Example

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

---

## openwrt_dhcp_host

Manages static DHCP host reservations.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `name` | String | Yes | Host name |
| `ip` | String | Yes | Reserved IP address |
| `mac` | String | Yes | MAC address |
| `leasetime` | String | No | Lease time override |
| `dns` | Bool | No | Add to DNS |
| `cloudflare` | Bool | No | Cloudflare DNS update |
| `ipv6_prefix` | String | No | IPv6 prefix |

### Example

```hcl
resource "openwrt_dhcp_host" "server" {
  name      = "fileserver"
  ip        = "192.168.1.100"
  mac       = "00:11:22:33:44:55"
  leasetime = "infinite"
}

resource "openwrt_dhcp_host" "printer" {
  name = "printer"
  ip   = "192.168.1.150"
  mac  = "AA:BB:CC:DD:EE:FF"
}
```

---

## openwrt_dhcp_odhcpd

Manages DHCPv6 (odhcpd) settings.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `maindhcp` | Bool | No | DHCPv4 on main interface |
| `leasefile` | String | No | Lease file path |
| `leasetrigger` | String | No | Script to run on lease changes |
| `loglevel` | Int64 | No | Log level (0-7) |

### Example

```hcl
resource "openwrt_dhcp_odhcpd" "main" {
  maindhcp     = false
  leasefile   = "/tmp/hosts/odhcpd"
  leasetrigger = "/usr/sbin/odhcpd-update"
  loglevel    = 4
}
```

---

## Import

DHCP resources can be imported using the config name and section name:

```bash
terraform import openwrt_dhcp_pool.lan lan
terraform import openwrt_dhcp_dnsmasq.main dnsmasq
terraform import openwrt_dhcp_host.server host
terraform import openwrt_dhcp_odhcpd.main odhcpd
```