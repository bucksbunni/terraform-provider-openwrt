---
layout: ""
page_title: "System Resources"
description: |-
  System resource types for the OpenWrt provider.
---

# System Resources

## openwrt_system

Manages system settings.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `hostname` | String | No | System hostname |
| `ttylogin` | Bool | No | Require TTY login |
| `log_size` | Int64 | No | Log buffer size (KB) |
| `urandom_seed` | Bool | No | Save random seed on shutdown |
| `zonename` | String | No | Timezone |
| `log_proto` | String | No | Log protocol: 'udp', 'file' |
| `conloglevel` | Int64 | No | Console log level |
| `cronloglevel` | Int64 | No | Cron log level |

### Example

```hcl
resource "openwrt_system" "main" {
  hostname     = "openwrt-router"
  ttylogin     = false
  log_size     = 128
  urandom_seed = true
  zonename     = "UTC"
  log_proto    = "udp"
  conloglevel  = 8
  cronloglevel = 7
}
```

---

## openwrt_system_ntp

Manages NTP (timeserver) configuration.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `name` | String | Yes | Config name |
| `enabled` | Bool | No | Enable NTP client |
| `server` | String | No | NTP servers (space-separated) |
| `use_dhcp` | Bool | No | Use servers from DHCP |

### Example

```hcl
resource "openwrt_system_ntp" "main" {
  name    = "ntp"
  enabled = true
  server  = "0.openwrt.pool.ntp.org 1.openwrt.pool.ntp.org"
}
```

---

## openwrt_system_led

Manages system LED configuration.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `name` | String | Yes | LED name |
| `sysfs` | String | No | LED device path |
| `trigger` | String | No | Trigger type |
| `mode` | String | No | LED mode (e.g., 'link tx rx') |
| `dev` | String | No | Network device |
| `default` | Bool | No | Default LED state |

### Example

```hcl
# Network activity LED
resource "openwrt_system_led" "wan_led" {
  name     = "WAN LED"
  sysfs   = "apu:green:3"
  trigger = "netdev"
  mode    = "link tx rx"
  dev     = "eth0"
}

# Static LED
resource "openwrt_system_led" "diag_led" {
  name     = "DIAG"
  sysfs   = "apu:green:1"
  trigger = "none"
  default = true
}
```

---

## openwrt_dropbear

Manages Dropbear SSH server settings.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `name` | String | Yes | Instance name |
| `password_auth` | Bool | No | Enable password authentication |
| `port` | Int64 | No | SSH port |
| `root_password_auth` | Bool | No | Password auth for root |
| `root_login` | Bool | No | Allow root login |

### Example

```hcl
resource "openwrt_dropbear" "main" {
  name             = "main"
  password_auth    = false
  port             = 22
  root_password_auth = true
  root_login       = true
}
```

---

## openwrt_uhttpd

Manages uHTTPd web server settings.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `name` | String | Yes | Instance name |
| `listen_http` | String | No | HTTP listeners |
| `listen_https` | String | No | HTTPS listeners |
| `redirect_https` | Bool | No | Redirect HTTP to HTTPS |
| `home` | String | No | Document root |
| `rfc1918_filter` | Bool | No | Filter private addresses |
| `max_requests` | Int64 | No | Max concurrent requests |
| `max_connections` | Int64 | No | Max connections |
| `cert` | String | No | HTTPS certificate path |
| `key` | String | No | HTTPS key path |

### Example

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

---

## openwrt_uhttpd_cert

Manages uHTTPd certificate generation settings.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `name` | String | Yes | Certificate name |
| `days` | Int64 | No | Validity days |
| `key_type` | String | No | Key type: 'rsa', 'ec' |
| `ec_curve` | String | No | EC curve (e.g., 'P-256') |
| `country` | String | No | Country code |
| `state` | String | No | State/province |
| `location` | String | No | City |
| `commonname` | String | No | Common name |

### Example

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

---

## openwrt_rpcd

Manages RPC daemon (ubus) settings.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `socket` | String | No | Unix socket path |
| `timeout` | Int64 | No | RPC timeout (seconds) |

### Example

```hcl
resource "openwrt_rpcd" "main" {
  socket  = "/var/run/ubus/ubus.sock"
  timeout = 30
}
```

---

## openwrt_sys_reboot

Triggers a reboot of the OpenWrt device.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID (sys_reboot) |
| `delay` | Int64 | No | Delay in seconds before rebooting (default: 0) |

### Example

```hcl
# Reboot immediately
resource "openwrt_sys_reboot" "now" {}

# Reboot after delay
resource "openwrt_sys_reboot" "delayed" {
  delay    = 10
}
```

### Notes

- The `delay` attribute causes Terraform to wait before sending the reboot command
- After reboot, the resource will be removed from state and must be re-applied
- Use `terraform apply -refresh=false` after reboot to avoid errors

---

## openwrt_sys_modprobe

Loads or unloads a kernel module on the OpenWrt device.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID (sys_modprobe/<module_name>) |
| `name` | String | Yes | Name of the kernel module |
| `action` | String | No | Action: 'load' (default) or 'unload' |
| `param` | Map | No | Module parameters as key-value pairs |
| `output` | String | Computed | Output from modprobe command |

### Example

```hcl
# Load wireless driver module
resource "openwrt_sys_modprobe" "ath10k" {
  name   = "ath10k_pci"
  action = "load"
}

# Load module with parameters
resource "openwrt_sys_modprobe" "ath10k_debug" {
  name   = "ath10k_pci"
  action = "load"
  param = {
    debug = "2"
  }
}

# Unload module
resource "openwrt_sys_modprobe" "unload_ath10k" {
  name   = "ath10k_pci"
  action = "unload"
}
```

### Notes

- Removing the resource from config will attempt to unload the module
- The `Read` function checks if the module is loaded; if not, removes from state
- Use with `openwrt_ipkg_package` to ensure kernel module packages are installed first

---

## Import

System resources can be imported using the config name and section name:

```bash
terraform import openwrt_system.main system
terraform import openwrt_system_ntp.main ntp
terraform import openwrt_system_led.wan_led wan_led
terraform import openwrt_dropbear.main main
terraform import openwrt_uhttpd.main main
terraform import openwrt_uhttpd_cert.defaults defaults
terraform import openwrt_rpcd.main rpcd

# Sys resources
terraform import openwrt_sys_reboot.main sys_reboot
terraform import openwrt_sys_modprobe.ath10k sys_modprobe/ath10k_pci
```