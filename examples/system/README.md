# System Configuration Examples

This directory contains Terraform configurations for managing OpenWrt system settings, NTP, LEDs, and SSH access.

## Files

- `main.tf` - Complete system configuration with provider setup
- `provider.tf` - Reusable provider configuration with variables
- `README.md` - This file

## Resources Created

### System Settings

- `openwrt_system` for hostname, timezone, logging configuration

### NTP Client

- `openwrt_system_ntp` for time synchronization with pool servers

### System LEDs

- `openwrt_system_led` for status LEDs (WAN activity, WiFi, USB)
- Supports netdev, timer, and heartbeat triggers

### SSH Server (Dropbear)

- `openwrt_dropbear` for SSH access configuration

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
