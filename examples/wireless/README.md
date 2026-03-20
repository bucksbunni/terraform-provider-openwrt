# Wireless Configuration Examples

This directory contains Terraform configurations for managing OpenWrt wireless radios and SSIDs.

## Files

- `main.tf` - Complete wireless configuration with provider setup
- `provider.tf` - Reusable provider configuration with variables
- `README.md` - This file

## Resources Created

### Radio Devices

- `openwrt_wireless_device` for radio0 (2.4GHz) and radio1 (5GHz)
- Channel, HT mode, country code, and enable/disable settings

### WiFi Interfaces

- `openwrt_wireless_iface` for HomeSS (WPA2-PSK), GuestSS (isolated), and IoT-SSID
- MAC filtering, client isolation, and encryption settings

## Usage

```bash
terraform init
terraform plan
terraform apply
```

## Variables

Configure via `provider.tf` or environment variables:
- `openwrt_host` - LuCI URL (default: http://192.168.1.1)
- `openwrt_username` - SSH username (default: root)
- `openwrt_password` - SSH password
- `insecure` - Skip TLS verification (default: true)
