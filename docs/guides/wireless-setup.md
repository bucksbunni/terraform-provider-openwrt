---
layout: ""
page_title: "Wireless Setup Guide"
description: |-
  Guide for setting up wireless on OpenWrt with the Terraform provider.
---

# Wireless Setup Guide

Wireless support in OpenWrt requires kernel modules and firmware packages that are hardware-specific. These must be installed separately as they vary by chipset.

## Identifying Your Hardware

Check your wireless hardware with:

```sh
lspci | grep -i wireless
lsusb | grep -i wifi
```

Search for available packages:

```sh
opkg update
opkg list | grep -E 'kmod-|ath10k|brcmfmac|mt76'
```

## Common Chipsets

| Chipset | Kernel Module | Firmware Package |
|---------|--------------|------------------|
| Atheros (QCA988x) | `kmod-ath10k` | `ath10k-firmware-qca988x` |
| Atheros (QCA6174) | `kmod-ath10k` | `ath10k-firmware-qca6174` |
| Broadcom | `kmod-brcmfmac` | `brcmfmac-firmware` |
| MediaTek | `kmod-mt76` | `mt7601u-firmware` |
| Intel | `kmod-iwlwifi` | `iwlwifi-firmware` |

## Installing via Terraform

Use the `openwrt_ipkg_package` resource to install the required packages:

```hcl
# Install kernel module and firmware
resource "openwrt_ipkg_package" "ath10k_dkmod" {
  name = "kmod-ath10k"
}

resource "openwrt_ipkg_package" "ath10k_fw" {
  name = "ath10k-firmware-qca988x"
}

# Restart network to load the driver
resource "openwrt_sys_rpc" "wifi_restart" {
  method      = "init.restart"
  params_json = jsonencode(["network"])

  depends_on = [
    openwrt_ipkg_package.ath10k_dkmod,
    openwrt_ipkg_package.ath10k_fw
  ]
}

# Now configure wireless
resource "openwrt_wireless_device" "radio0" {
  name     = "radio0"
  type     = "mac80211"
  channel  = 6
  htmode   = "HT40"
  disabled = false
}

resource "openwrt_wireless_iface" "main" {
  name       = "testwifi"
  device     = "radio0"
  mode       = "ap"
  ssid       = "MyNetwork"
  encryption = "psk2"
  key        = "secretpassword"
  network    = "lan"
}
```

## Driver Loading

After installing the packages, the network subsystem must be restarted to load the driver. The `init.restart network` call above handles this.

Alternatively, you can use:

```hcl
resource "openwrt_sys_rpc" "wifi_up" {
  method      = "sys.exec"
  params_json = jsonencode(["wifi up"])
}
```

## Manual Verification

If wireless doesn't appear, verify:

1. Package installed: `opkg list-installed | grep kmod-ath`
2. Driver loaded: `lsmod | grep ath`
3. Interface exists: `ip link show`
4. Check dmesg: `dmesg | grep ath`

## Troubleshooting

- **No wireless interfaces appear**: Kernel module may not support your device. Check OpenWrt compatibility lists.
- **Firmware missing**: Ensure the firmware package is installed (e.g., `ath10k-firmware-qca988x`)
- **Driver fails to load**: Some devices require specific firmware versions or additional packages