---
layout: ""
page_title: "Network Data Sources"
description: |-
  Network data source types for the OpenWrt provider.
---

# Network Data Sources

## openwrt_sys_net_devices

Retrieves network device information.

### Schema

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Always 'net_devices' |
| `devices` | List | List of network device objects |

Device object:
- `name` - Device name
- `type` - Device type
- `up` - Whether up
- `mtu` - MTU
- `mac` - MAC address
- `txqueuelen` - TX queue length
- `statistics` - Traffic stats (bytes, packets, errors) |

### Example

```hcl
data "openwrt_sys_net_devices" "main" {}

output "device_names" {
  value = [for d in data.openwrt_sys_net_devices.main.devices : d.name]
}

output "eth0_mac" {
  value = [for d in data.openwrt_sys_net_devices.main.devices : d.mac if d.name == "eth0"][0]
}
```

---

## openwrt_sys_net_routes

Retrieves the IPv4 routing table.

### Schema

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Always 'net_routes' |
| `routes` | List | List of route objects |

Route object:
- `target` - Destination network
- `netmask` - Destination mask
- `gateway` - Next hop
- `metric` - Route metric
- `device` - Output device |
- `table` - Routing table |

### Example

```hcl
data "openwrt_sys_net_routes" "main" {}

output "default_gateway" {
  value = [for r in data.openwrt_sys_net_routes.main.routes : r.gateway if r.target == "0.0.0.0"][0]
}
```

---

## openwrt_sys_net_routes6

Retrieves the IPv6 routing table.

### Schema

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Always 'net_routes6' |
| `routes` | List | List of IPv6 route objects |

### Example

```hcl
data "openwrt_sys_net_routes6" "main" {}

output "ipv6_routes" {
  value = data.openwrt_sys_net_routes6.main.routes
}
```

---

## openwrt_sys_net_arptable

Retrieves the ARP table.

### Schema

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Always 'net_arptable' |
| `entries` | List | List of ARP entries |

ARP entry:
- `ip` - IP address
- `mac` - MAC address
- `device` - Network device |

### Example

```hcl
data "openwrt_sys_net_arptable" "main" {}

output "arp_count" {
  value = length(data.openwrt_sys_net_arptable.main.entries)
}

output "gateway_mac" {
  value = [for e in data.openwrt_sys_net_arptable.main.entries : e.mac if e.ip == "192.168.1.1"][0]
}
```

---

## openwrt_sys_net_conntrack

Retrieves the connection tracking table.

### Schema

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Always 'net_conntrack' |
| `entries` | List | List of connection tracking entries |

Entry object:
- `source` - Source IP:port
- `dest` - Destination IP:port
- `protocol` - Protocol number/name
- `state` - Connection state |
- `timeout` - Timeout |

### Example

```hcl
data "openwrt_sys_net_conntrack" "main" {}

output "tcp_connections" {
  value = length([for c in data.openwrt_sys_net_conntrack.main.entries : c if c.protocol == "6"])
}
```

---

## Usage with openwrt_sys_rpc

For network data not covered by dedicated data sources, use `openwrt_sys_rpc`:

```hcl
# Get detailed device statistics
data "openwrt_sys_rpc" "device_stats" {
  method      = "net.deviceinfo"
  params_json = "[]"
}

# Get WiFi scan results (if supported)
data "openwrt_sys_rpc" "wifi_scan" {
  method      = "wifi.getiwinfo"
  params_json = jsonencode(["wlan0"])
}
```