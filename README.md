# Terraform Provider for OpenWrt (JSON‑RPC)

This provider manages OpenWrt devices via the LuCI JSON‑RPC API.

It wraps the following LuCI RPC libraries:

- `/cgi-bin/luci/rpc/uci` – UCI configuration
- `/cgi-bin/luci/rpc/fs` – filesystem operations
- `/cgi-bin/luci/rpc/sys` – system info and utilities
- `/cgi-bin/luci/rpc/ipkg` – package manager (opkg)

The upstream JSON‑RPC behaviour is documented in `JsonRpcHowTo.md` (included in this repository) and the LuCI LuaDocs under `docs/`.

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

Place `terraform-provider-openwrt` in your Terraform plugin directory or use terraform init with `source = "bucksbunni/openwrt"` once you publish the provider.

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

### openwrt_uci_section

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

### openwrt_fs_file

Manage a file via `/rpc/fs`:

```hcl
resource "openwrt_fs_file" "motd" {
  path    = "/etc/motd"
  content = "Managed by Terraform\n"
}
```

Content is base64‑encoded on the wire; the resource works with plain UTF‑8.

### openwrt_ipkg_package

Ensure a package is installed using `/rpc/ipkg`:

```hcl
resource "openwrt_ipkg_package" "luci_mod_rpc" {
  name = "luci-mod-rpc"
}
```

### openwrt_sys_rpc (data source)

Low‑level `/rpc/sys` access:

```hcl
data "openwrt_sys_rpc" "hostname" {
  method = "hostname"
}

output "hostname" {
  value = jsondecode(data.openwrt_sys_rpc.hostname.result_json)
}
```

Example: IPv4 routes:

```hcl
data "openwrt_sys_rpc" "routes" {
  method      = "net.routes"
  params_json = "[]"
}

locals {
  routes = jsondecode(data.openwrt_sys_rpc.routes.result_json)
}

output "routes" {
  value = local.routes
}
```

## Limitations and TODOs

- Only a subset of UCI and ipkg RPC functions are wrapped.
- /rpc/sys is exposed generically; for convenience, typed data sources (e.g. openwrt_sys_hostname, openwrt_sys_routes) can be added.
- LuCI path is fixed to /cgi-bin/luci; making it configurable is a future enhancement.
- TLS options are limited to insecure; supporting CA bundles and client certs would be useful for production.

## License

This project has been developed under the [GNU GPLv3](./LICENSE) license.
