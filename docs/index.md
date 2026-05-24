---
layout: ""
page_title: "Provider: OpenWrt"
description: |-
  The OpenWrt provider manages OpenWrt devices via the LuCI JSON-RPC API.
---

# OpenWrt Provider

The OpenWrt provider allows you to manage OpenWrt devices via the LuCI JSON-RPC API. It wraps the following LuCI RPC libraries:

- `/cgi-bin/luci/rpc/uci` – UCI configuration
- `/cgi-bin/luci/rpc/fs` – filesystem operations
- `/cgi-bin/luci/rpc/sys` – system info and utilities
- `/cgi-bin/luci/rpc/ipkg` – package manager (opkg)

## Prerequisites

On the router:

```sh
opkg update
opkg install luci-mod-rpc luci-lib-ipkg luci-compat
/etc/init.d/uhttpd restart
```

Ensure LuCI is reachable at something like:

- http://192.168.1.1/cgi-bin/luci
- https://router.example/cgi-bin/luci

## Quick Start

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
  insecure = true
}
```

## Guides

- [Wireless Setup Guide](guides/wireless-setup.md) - Installing wireless drivers and configuring WiFi

## Resources

- [Network Resources](resources/network.md) - Interfaces, devices, bridges, VLANs, WireGuard
- [Firewall Resources](resources/firewall.md) - Zones, rules, forwarding
- [DHCP Resources](resources/dhcp.md) - DHCP pools, dnsmasq, static reservations
- [Wireless Resources](resources/wireless.md) - Radio devices and SSIDs (see examples/wireless)
- [System Resources](resources/system.md) - System settings, NTP, LEDs
- [Generic Resources](resources/generic.md) - UCI sections, files, packages

## Data Sources

- [System Data Sources](data-sources/sys.md) - Hostname, uptime, processes
- [Network Data Sources](data-sources/network.md) - Devices, routes, ARP table

## Environment Variables

The provider supports reading configuration from environment variables:

- `OPENWRT_HOST` - OpenWrt host URL
- `OPENWRT_USERNAME` - Username for authentication
- `OPENWRT_PASSWORD` - Password for authentication
- `OPENWRT_INSECURE` - Set to "true" to skip TLS verification

-> **Note:** Environment variables are useful for CI/CD pipelines to avoid exposing credentials in configuration files.