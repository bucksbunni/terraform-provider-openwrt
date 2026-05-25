---
layout: ""
page_title: "Generic Resources"
description: |-
  Generic resource types for the OpenWrt provider.
---

# Generic Resources

## openwrt_uci_section

Manages a generic UCI section and its options.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `config` | String | Yes | UCI config file (e.g., 'network', 'firewall') |
| `type` | String | Yes | Section type (e.g., 'interface', 'zone', 'rule') |
| `name` | String | No | Named section name |
| `options` | Map | No | UCI options as key-value pairs |

### Example

```hcl
# Simple interface
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

# Anonymous section
resource "openwrt_uci_section" "static_route" {
  config = "network"
  type   = "route"
  name   = ""

  options = {
    interface = "wan"
    target    = "10.0.0.0/8"
    gateway   = "192.168.1.1"
  }
}
```

---

## openwrt_fs_file

Manages files on the router filesystem.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `path` | String | Yes | File path |
| `content` | String | No | File content (plain UTF-8) |
| `mode` | String | No | File permissions (e.g., '0644') |
| `user` | String | No | File owner |
| `group` | String | No | File group |
| `dir_mode` | String | No | Directory permissions |

### Example

```hcl
# Create a file
resource "openwrt_fs_file" "motd" {
  path    = "/etc/motd"
  content = "Welcome to OpenWrt!\n"
  mode    = "0644"
}

# Create a configuration file
resource "openwrt_fs_file" "custom_config" {
  path    = "/etc/config/custom"
  content = <<-EOF
config custom 'settings'
    option enabled '1'
    option debug '0'
EOF
  mode    = "0644"
}
```

---

## openwrt_ipkg_package

Manages package installation/removal via opkg.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID (equals package name) |
| `name` | String | Yes | Package name as known to opkg |
| `autoremove` | Bool | No | Remove packages that were installed automatically to satisfy dependencies (default: true) |
| `force_remove` | Bool | No | Remove package and all dependencies (default: true) |
| `update` | Bool | No | Update package lists before installing (default: true) |

### Example

```hcl
# Install LuCI RPC module
resource "openwrt_ipkg_package" "luci_mod_rpc" {
  name = "luci-mod-rpc"
}

# Install wireless driver (all flags default to true)
resource "openwrt_ipkg_package" "ath10k" {
  name = "kmod-ath10k"
}

# Install with explicit flags
resource "openwrt_ipkg_package" "ath10k_fw" {
  name         = "ath10k-firmware-qca988x"
  autoremove   = false
  force_remove = false
  update       = false  # skip update if already done
}

# Install firewall utility
resource "openwrt_ipkg_package" "iptables" {
  name = "iptables"
}
```

### Notes

- The package is installed during resource creation
- Removing the resource will uninstall the package
- Use this for managing packages that need to be present for other resources

---

## Import

Generic resources can be imported:

```bash
# UCI section - config.type.name format
terraform import openwrt_uci_section.lan network.interface.lan
terraform import openwrt_uci_section.route network.route.myroute

# Files - use the path
terraform import openwrt_fs_file.motd /etc/motd

# Packages - use the package name
terraform import openwrt_ipkg_package.luci_mod_rpc luci-mod-rpc
```