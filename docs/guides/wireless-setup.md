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
resource "openwrt_ipkg_package" "ath10k_kmod" {
  name = "kmod-ath10k"
}

resource "openwrt_ipkg_package" "ath10k_fw" {
  name = "ath10k-firmware-qca988x"
}

# Load the kernel module (required after package installation)
resource "openwrt_sys_modprobe" "ath10k" {
  name   = "ath10k_pci"
  action = "load"

  depends_on = [
    openwrt_ipkg_package.ath10k_kmod,
    openwrt_ipkg_package.ath10k_fw
  ]
}

# Alternative: Reboot to load driver automatically
# resource "openwrt_sys_reboot" "reload" {
#   delay = 5
# }

# Verify wireless is available
data "openwrt_sys_wireless_info" "status" {}

# Now configure wireless
resource "openwrt_wireless_device" "radio0" {
  name     = "radio0"
  type     = "mac80211"
  channel  = 6
  htmode   = "HT40"
  disabled = false

  depends_on = [openwrt_sys_modprobe.ath10k]
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

After installing the kernel module packages, you must either:

1. **Use `openwrt_sys_modprobe`** to load the driver immediately (recommended)
2. **Use `openwrt_sys_reboot`** to reboot the device so the module loads automatically

The kernel module must be loaded before the wireless radio will appear in UCI or show up in `ip link`.

## Manual Verification

If wireless doesn't appear, verify:

1. Package installed: `opkg list-installed | grep kmod-ath`
2. Driver loaded: `lsmod | grep ath`
3. Interface exists: `ip link show`
4. Check dmesg: `dmesg | grep ath`

## Troubleshooting

- **No wireless interfaces appear**: Kernel module may not be loaded. Use `openwrt_sys_modprobe` to load it, or `openwrt_sys_reboot` to reboot.
- **Firmware missing**: Ensure the firmware package is installed (e.g., `ath10k-firmware-qca988x`)
- **Driver fails to load**: Some devices require specific firmware versions or additional packages
- **Wrong firmware**: Some devices (e.g., QCA9887) need different firmware than similar models (e.g., QCA988X). Check `lspci` output for exact device ID.