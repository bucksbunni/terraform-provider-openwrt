# Wireless Configuration Examples

This directory contains Terraform configurations for managing OpenWrt wireless radios and SSIDs.

## Prerequisites

Wireless support requires kernel modules and firmware packages to be installed on the router. These are hardware-specific and must be installed separately.

### Common Chipsets

| Chipset | Kernel Module | Firmware Package |
|---------|--------------|------------------|
| Atheros (QCA988x) | `kmod-ath10k` | `ath10k-firmware-qca988x` |
| Atheros (QCA6174) | `kmod-ath10k` | `ath10k-firmware-qca6174` |
| Broadcom | `kmod-brcmfmac` | `brcmfmac-firmware` |
| MediaTek | `kmod-mt76` | `mt7601u-firmware` |
| Intel | `kmod-iwlwifi` | `iwlwifi-firmware` |

Check your hardware with `lspci` or `lsusb` and search for packages with `opkg find kmod-* ath*`.

### Installation via Terraform

Use `openwrt_ipkg_package` to install the required packages:

```hcl
resource "openwrt_ipkg_package" "ath10k_dkmod" {
  name = "kmod-ath10k"
}

resource "openwrt_ipkg_package" "ath10k_fw" {
  name = "ath10k-firmware-qca988x"
}

resource "openwrt_sys_rpc" "wifi_restart" {
  method      = "init.restart"
  params_json = jsonencode(["network"])

  depends_on = [
    openwrt_ipkg_package.ath10k_dkmod,
    openwrt_ipkg_package.ath10k_fw
  ]
}
```

After installing packages, the network subsystem must be restarted to load the driver. Use `openwrt_sys_rpc` with `init.restart network` or call `wifi up`.

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
