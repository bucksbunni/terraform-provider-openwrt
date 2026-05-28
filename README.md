# Terraform Provider OpenWrt

The OpenWrt provider manages OpenWrt devices via the LuCI JSON-RPC API.

It wraps the following LuCI RPC libraries:

- `/cgi-bin/luci/rpc/uci` – UCI configuration
- `/cgi-bin/luci/rpc/fs` – filesystem operations
- `/cgi-bin/luci/rpc/sys` – system info and utilities
- `/cgi-bin/luci/rpc/ipkg` – package manager (opkg)

The upstream JSON‑RPC behaviour is documented in `JsonRpcHowTo.md` of the [LuCI Wiki](https://github.com/openwrt/luci/wiki).

> **WARNING: All features are highly experimental!** Use in production at your own risk.

Full documentation is available in the [docs](docs/index.md) directory. Ready-to-use configurations in [examples](examples/).

## Requirements

Enable the LuCI RPC interface on your OpenWrt router:

```sh
opkg update
opkg install luci-mod-rpc luci-lib-ipkg luci-compat
/etc/init.d/uhttpd restart
```

Ensure LuCI is reachable at something like:

- http://192.168.1.1/cgi-bin/luci
- or https://router.example/cgi-bin/luci

## Installation

The provider is published on the [Terraform Registry](https://registry.terraform.io/providers/bucksbunni/openwrt). Add it to your Terraform configuration:

```hcl
terraform {
  required_providers {
    openwrt = {
      source  = "bucksbunni/openwrt"
      version = "0.1.0"
    }
  }
}
```

Then run `terraform init` to download the provider.

## Provider configuration

```hcl
provider "openwrt" {
  host     = "http://192.168.1.1"
  username = "root"
  password = "yourpassword"
  insecure = true # only if using self-signed HTTPS
}

# Example: Configure a network interface
resource "openwrt_network_interface" "lan" {
  name    = "lan"
  proto   = "static"
  device  = "br-lan"
  ipaddr  = ["192.168.1.1/24"]
}
```

`host` must point at the LuCI base URL (without `/rpc/...` appended).

## Limitations and TODOs

- LuCI path is fixed to /cgi-bin/luci; making it configurable is a future enhancement.
- TLS options are limited to insecure; supporting CA bundles and client certs would be useful for production.
- Acceptance tests require a live OpenWrt device.
- Import support could be extended for additional resources.

## License

This project is licensed under the Mozilla Public License 2.0 (MPL 2.0) - see [LICENSE](LICENSE) for details.