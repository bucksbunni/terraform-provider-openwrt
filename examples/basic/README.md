# Basic Provider Configuration

This example demonstrates the minimal configuration needed to connect Terraform to an OpenWrt device.

## Prerequisites

On the OpenWrt router:

```sh
opkg update
opkg install luci-mod-rpc luci-lib-ipkg luci-compat
/etc/init.d/uhttpd restart
```

## Files

`main.tf` - Terraform configuration with provider and basic data sources

## Usage

```bash
terraform init
terraform plan
terraform apply
```

## Notes

- The provider connects to LuCI's JSON-RPC API at `/cgi-bin/luci/rpc`
- Authentication uses the router's root credentials
- For production use, consider:
  - Using environment variables for credentials
  - Enabling HTTPS on the router
  - Using aCA bundle instead of `insecure = true`
