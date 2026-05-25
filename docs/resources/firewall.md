---
layout: ""
page_title: "Firewall Resources"
description: |-
  Firewall resource types for the OpenWrt provider.
---

# Firewall Resources

## openwrt_firewall_zone

Manages firewall zones.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `name` | String | Yes | Zone name |
| `input` | String | No | Input policy: 'ACCEPT', 'REJECT', 'DROP' |
| `output` | String | No | Output policy: 'ACCEPT', 'REJECT', 'DROP' |
| `forward` | String | No | Forward policy: 'ACCEPT', 'REJECT', 'DROP' |
| `masq` | Bool | No | Enable SNAT masquerading |
| `masq6` | Bool | No | Enable IPv6 masquerading |
| `network` | List(String) | No | Networks assigned to zone (e.g., ['lan', 'guest']) |

### Example

```hcl
resource "openwrt_firewall_zone" "lan" {
  name     = "lan"
  input    = "ACCEPT"
  output   = "ACCEPT"
  forward  = "REJECT"
  masq     = false
  network  = ["lan"]
}

resource "openwrt_firewall_zone" "guest" {
  name     = "guest"
  input    = "REJECT"
  output   = "ACCEPT"
  forward  = "REJECT"
  masq     = true
  network  = ["guest", "iot"]
}
```

---

## openwrt_firewall_rule

Manages firewall rules.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `name` | String | Yes | Rule name |
| `src` | String | No | Source zone |
| `dest` | String | No | Destination zone |
| `proto` | String | No | Protocol: 'tcp', 'udp', 'tcpudp', 'icmp', 'all' |
| `src_port` | String | No | Source port or range |
| `dest_port` | String | No | Destination port or range |
| `src_ip` | String | No | Source IP/CIDR |
| `dest_ip` | String | No | Destination IP/CIDR |
| `target` | String | Yes | Target: 'ACCEPT', 'REJECT', 'DROP' |
| `family` | String | No | IP family: 'ipv4', 'ipv6', 'all' |
| `icmp_type` | List(String) | No | ICMP type (e.g., ['echo-request', 'echo-reply']) |
| `limit` | String | No | Rate limit (e.g., '10/minute') |
| `extra` | String | No | Extra iptables options |
| `enabled` | Bool | No | Enable the rule |

### Example

```hcl
# Allow DNS from guest to lan
resource "openwrt_firewall_rule" "allow_dns" {
  name      = "Allow-DNS"
  src       = "guest"
  dest      = "lan"
  dest_port = "53"
  proto     = "tcpudp"
  target    = "ACCEPT"
}

# Allow HTTP/HTTPS from wan to lan
resource "openwrt_firewall_rule" "allow_http" {
  name        = "Allow-HTTP"
  src         = "wan"
  dest        = "lan"
  dest_port   = "80"
  proto       = "tcp"
  target      = "ACCEPT"
  src_ip      = "192.168.1.0/24"
}

# Rate-limited ping
resource "openwrt_firewall_rule" "allow_ping" {
  name      = "Allow-Ping"
  src       = "wan"
  proto     = "icmp"
  target    = "ACCEPT"
  limit     = "5/second"
}
```

---

## openwrt_firewall_forwarding

Manages zone-to-zone traffic forwarding.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `src` | String | Yes | Source zone |
| `dest` | String | Yes | Destination zone |
| `enabled` | Bool | No | Enable forwarding |

### Example

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

---

## openwrt_firewall_defaults

Manages firewall default policies.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `input` | String | No | Default input policy |
| `output` | String | No | Default output policy |
| `forward` | String | No | Default forward policy |
| `synflood_protect` | Bool | No | Enable SYN flood protection |
| `synflood_rate` | String | No | SYN flood rate limit |
| `drop_invalid` | Bool | No | Drop invalid packets |
| `auto_helper` | Bool | No | Auto-load helper modules |

### Example

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

---

## Import

Firewall resources can be imported using the config name and section name:

```bash
terraform import openwrt_firewall_zone.lan lan
terraform import openwrt_firewall_rule.allow_dns allow_dns
terraform import openwrt_firewall_forwarding.lan_to_wan lan_wan
terraform import openwrt_firewall_defaults.main defaults
```