---
layout: ""
page_title: "Network Resources"
description: |-
  Network resource types for the OpenWrt provider.
---

# Network Resources

## openwrt_network_interface

Manages a network interface (LAN, WAN, guest, etc.).

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID: network/<interface_name> |
| `name` | String | Yes | Interface name (e.g., 'lan', 'wan', 'guest') |
| `proto` | String | Yes | Protocol: 'static', 'dhcp', 'dhcpv6', 'pppoe', 'wireguard', 'qmi' |
| `device` | String | No | Network device (e.g., 'eth0', 'br-lan', 'wg0') |
| `ipaddr` | String | No | IPv4 address with prefix (e.g., '192.168.1.1/24') |
| `netmask` | String | No | Netmask for static IP |
| `gateway` | String | No | Default gateway |
| `dns` | String | No | DNS servers (space-separated) |
| `metric` | Int64 | No | Interface metric |
| `delegate` | Bool | No | Delegate IPv6 prefixes |
| `ip6addr` | String | No | IPv6 address |
| `ip6prefix` | String | No | IPv6 prefix |
| `ip6assign` | String | No | IPv6 assignment prefix |
| `ip6gateway` | String | No | IPv6 gateway |
| `auto` | String | No | Auto-enable on boot ('1' or '0') |
| `type` | String | No | Interface type |
| `bridge_empty` | Bool | No | Create empty bridge |

### Example

```hcl
# Static LAN interface
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

# DHCP client WAN interface
resource "openwrt_network_interface" "wan" {
  name    = "wan"
  proto   = "dhcp"
  device  = "eth0"
  metric  = 200
  auto    = "1"
}
```

---

## openwrt_network_device

Manages network devices (bridges, bonds, VLANs).

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `name` | String | Yes | Device name (e.g., 'br-lan') |
| `type` | String | Yes | Device type: 'bridge', 'bonding', 'vlan' |
| `enabled` | Bool | No | Enable the device |
| ` Bridge Ports | ports` | String | No | Bridge ports (space-separated) |
| ` Bridge Empty | bridge_empty` | Bool | No | Create empty bridge |
| ` Bonding | policy` | String | No | Bonding policy (e.g., '802.3ad') |
| ` Bonding | xmit_hash_policy` | String | No | Transmission hash policy |
| ` VLAN | tag` | Int64 | No | VLAN tag |
| ` VLAN | device` | String | No | Parent device |

### Example

```hcl
# Bridge device
resource "openwrt_network_device" "br_lan" {
  name  = "br-lan"
  type  = "bridge"
  ports = "eth0 eth1"
}

# Bonding device
resource "openwrt_network_device" "bond0" {
  name            = "bond0"
  type            = "bonding"
  ports           = "eth1 eth2"
  policy          = "802.3ad"
  xmit_hash_policy = "layer2+3"
}
```

---

## openwrt_network_bridge_vlan

Manages bridge VLAN assignments.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `device` | String | Yes | Bridge device name |
| `vlan` | Int64 | Yes | VLAN ID |
| `ports` | String | Yes | Ports with flags (e.g., 'eth0:u* eth1:t') |

### Example

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

---

## openwrt_network_globals

Manages network global settings.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `ula_prefix` | String | No | IPv6 ULA prefix |
| `packet_steering` | Bool | No | Enable packet steering |

### Example

```hcl
resource "openwrt_network_globals" "main" {
  ula_prefix     = "fda1:b7f5:46f7::/48"
  packet_steering = true
}
```

---

## openwrt_network_wireguard

Manages WireGuard VPN peers.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `name` | String | Yes | Interface name |
| `description` | String | No | Description |
| `public_key` | String | Yes | Own public key |
| `private_key` | String | No | Own private key |
| `listen_port` | Int64 | No | Listen port |
| `address` | String | No | Interface addresses (comma-separated) |
| `dns` | String | No | DNS servers |
| `mtu` | Int64 | No | MTU |
| `endpoint_host` | String | No | Endpoint host |
| `endpoint_port` | Int64 | No | Endpoint port |
| `allowed_ips` | String | No | Allowed IPs |
| `persistent_keepalive` | Int64 | No | Keepalive interval (seconds) |

### Example

```hcl
resource "openwrt_network_wireguard" "wg0" {
  name         = "wg0"
  description  = "VPN interface"
  listen_port  = 51820
  address      = "10.0.0.1/24"
  private_key  = "cB7D9F3A2E1..."
}

resource "openwrt_network_wireguard" "peer1" {
  name                  = "peer1"
  description           = "Remote peer 1"
  public_key            = "ABC123...="
  endpoint_host         = "vpn.example.com"
  endpoint_port         = 51820
  persistent_keepalive  = 25
  allowed_ips           = "10.0.0.2/32 10.0.0.0/24"
}
```

---

## Import

Network resources can be imported using the config name and section name:

```bash
terraform import openwrt_network_interface.lan lan
terraform import openwrt_network_device.br_lan br-lan
terraform import openwrt_network_bridge_vlan.vlan1 br-lan.1
terraform import openwrt_network_globals.main globals
terraform import openwrt_network_wireguard.wg0 wg0
```