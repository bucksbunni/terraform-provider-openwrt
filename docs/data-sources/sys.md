---
layout: ""
page_title: "System Data Sources"
description: |-
  System data source types for the OpenWrt provider.
---

# System Data Sources

## openwrt_sys_hostname

Retrieves the system hostname.

### Schema

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Always 'hostname' |
| `hostname` | String | System hostname |

### Example

```hcl
data "openwrt_sys_hostname" "main" {}

output "hostname" {
  value = data.openwrt_sys_hostname.main.hostname
}
```

---

## openwrt_sys_uptime

Retrieves the system uptime.

### Schema

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Always 'uptime' |
| `uptime` | Int64 | Uptime in seconds |

### Example

```hcl
data "openwrt_sys_uptime" "main" {}

output "uptime_seconds" {
  value = data.openwrt_sys_uptime.main.uptime
}
```

---

## openwrt_sys_process_list

Retrieves the list of running processes.

### Schema

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Always 'process_list' |
| `processes` | List | List of process objects |

Process object:
- `pid` - Process ID
- `name` - Process name
- `user` - Running user
- `stat` - Process state
- `cpu` - CPU usage
- `mem` - Memory usage

### Example

```hcl
data "openwrt_sys_process_list" "main" {}

output "process_count" {
  value = length(data.openwrt_sys_process_list.main.processes)
}
```

---

## openwrt_sys_init

Retrieves the list of init scripts and their status.

### Schema

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Always 'init' |
| `scripts` | List | List of init script objects |

Script object:
- `name` - Script name
- `enabled` - Whether enabled
- `running` - Whether running |

### Example

```hcl
data "openwrt_sys_init" "main" {}

output "enabled_services" {
  value = [for s in data.openwrt_sys_init.main.scripts : s.name if s.enabled]
}
```

---

## openwrt_sys_wireless

Retrieves wireless interface information.

### Schema

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Interface name |
| `ifname` | String | Interface name |
| `device` | String | Radio device |
| `mode` | String | Operating mode |
| `ssid` | String | SSID |
| `bssid` | String | BSSID |
| `channel` | Int64 | Current channel |
| `signal` | Int64 | Signal strength |
| `noise` | Int64 | Noise level |
| `txpower` | Int64 | Transmit power |
| `quality` | Int64 | Link quality |
| `quality_max` | Int64 | Max quality |

### Example

```hcl
data "openwrt_sys_wireless" "wlan0" {
  ifname = "wlan0"
}

output "signal_strength" {
  value = data.openwrt_sys_wireless.wlan0.signal
}
```

---

## openwrt_sys_wireless_radios

Discovers available wireless radio devices from OpenWrt's UCI configuration.

### Schema

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Always 'wireless_radios' |
| `radios` | List | List of wireless radio objects |

#### radios Object

| Attribute | Type | Description |
|-----------|------|-------------|
| `name` | String | Radio name (e.g., radio0) |
| `type` | String | Driver type (mac80211, broadcom) |
| `path` | String | Hardware path |
| `band` | String | Supported band (2g, 5g, 6g) |
| `channel` | Int64 | Current channel |
| `htmode` | String | Channel width (HT40, VHT80, etc.) |
| `country` | String | Country code |
| `txpower` | Int64 | Transmit power |
| `disabled` | Bool | Whether radio is disabled |

### Example

```hcl
data "openwrt_sys_wireless_radios" "available" {}

output "radio_count" {
  value = length(data.openwrt_sys_wireless_radios.available.radios)
}
```

---

## openwrt_sys_wireless_ifaces

Discovers wireless interfaces from OpenWrt's UCI configuration.

### Schema

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Always 'wireless_ifaces' |
| `interfaces` | List | List of wireless interface objects |

#### interfaces Object

| Attribute | Type | Description |
|-----------|------|-------------|
| `name` | String | Interface name (e.g., wifinet0) |
| `device` | String | Parent radio device |
| `mode` | String | Operation mode (ap, sta, adhoc, etc.) |
| `ssid` | String | Network SSID |
| `encryption` | String | Encryption mode |
| `network` | String | Attached network |
| `disabled` | Bool | Whether interface is disabled |
| `hidden` | Bool | Whether SSID is hidden |
| `macfilter` | String | MAC filter mode |
| `maclist` | String | MAC address list |
| `isolate` | Bool | Client isolation enabled |

### Example

```hcl
data "openwrt_sys_wireless_ifaces" "existing" {}

output "interface_count" {
  value = length(data.openwrt_sys_wireless_ifaces.interfaces)
}
```

---

## openwrt_sys_rpc

Low-level access to the /rpc/sys JSON-RPC API.

### Schema

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Method name |
| `method` | String | RPC method name |
| `params_json` | String | JSON-encoded parameters |
| `result_json` | String | Raw JSON result |

### Example

```hcl
# Get system info
data "openwrt_sys_rpc" "sysinfo" {
  method = "sysinfo"
}

# Parse result
output "sysinfo" {
  value = jsondecode(data.openwrt_sys_rpc.sysinfo.result_json)
}

# List wireless devices
data "openwrt_sys_rpc" "wifi_devices" {
  method      = "wifi.getiwinfo"
  params_json = "[]"
}

# Check if service is enabled
data "openwrt_sys_rpc" "network_enabled" {
  method      = "init.enabled"
  params_json = jsonencode(["network"])
}
```

### Available Methods

Common LuCI sys methods:
- `hostname` - Get hostname
- `uptime` - Get uptime
- `sysinfo` - Get system info
- `process.list` - List processes
- `net.deviceinfo` - Network device info
- `net.routes` - IPv4 routes
- `net.routes6` - IPv6 routes
- `net.conntrack` - Connection tracking
- `net.arptable` - ARP table
- `wifi.getiwinfo` - Wireless info
- `init.list` - List init scripts
- `init.enabled` - Check if service enabled
- `user.getuser` - Get user info