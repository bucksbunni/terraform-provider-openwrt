---
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

-> **Important:** The first time you install kernel modules, you must:
> 1. Ensure package lists are up to date (run `opkg update` manually first, or use the ipkg_package resource with the implicit update flag)
> 2. Install the kernel module package via Terraform
> 3. Add the module to `/etc/modules.d/` using `openwrt_sys_modules`
> 4. **Reboot the device** using `openwrt_sys_reboot` - kernel modules require a reboot to load properly

Use the `openwrt_ipkg_package` resource to install the required packages:

```hcl
# Install kernel module and firmware
resource "openwrt_ipkg_package" "ath10k_kmod" {
  name = "kmod-ath10k"
}

resource "openwrt_ipkg_package" "ath10k_fw" {
  name = "ath10k-firmware-qca988x"
}

# Configure modules to load at boot time (recommended for reliability)
resource "openwrt_sys_modules" "boot_modules" {
  modules = ["ath10k_pci"]

  depends_on = [
    openwrt_ipkg_package.ath10k_kmod,
    openwrt_ipkg_package.ath10k_fw
  ]
}

# Reboot to load modules at boot (required after first apply)
# After reboot, the wireless driver will be loaded before network services start
resource "openwrt_sys_reboot" "reload" {
  delay = 5

  depends_on = [openwrt_sys_modules.boot_modules]
}

# Verify wireless is available (after reboot)
data "openwrt_sys_wireless_radios" "available" {}

# Now configure wireless (after modules are loaded at boot)
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

After installing the kernel module packages, you must ensure the driver is loaded before wireless will work:

### Option 1: Boot-time loading (Recommended)

Use `openwrt_sys_modules` to configure modules to load automatically at boot:

```hcl
resource "openwrt_sys_modules" "boot_modules" {
  modules = ["ath10k_pci"]
}
```

This is more reliable than runtime loading because the module loads before network services start, avoiding race conditions that can cause `HOSTAPD_START_FAILED` errors.

After applying, reboot the device (manually or via `openwrt_sys_reboot`) for the changes to take effect.

### Option 2: Runtime loading

Use `openwrt_sys_modprobe` to load the driver immediately:

```hcl
resource "openwrt_sys_modprobe" "ath10k" {
  name   = "ath10k_pci"
  action = "load"
}
```

Note: This can cause race conditions where netifd tries to start hostapd before the driver is fully initialized.

## Manual Verification

If wireless doesn't appear, verify:

1. Package installed: `opkg list-installed | grep kmod-ath`
2. Driver loaded: `lsmod | grep ath`
3. Interface exists: `ip link show`
4. Check dmesg: `dmesg | grep ath`

## Troubleshooting

- **No wireless interfaces appear**: Use `openwrt_sys_modules` to configure boot-time loading, then reboot.
- **HOSTAPD_START_FAILED**: This usually means the driver loaded too late. Use `openwrt_sys_modules` instead of `openwrt_sys_modprobe` for boot-time loading.
- **Firmware missing**: Ensure the firmware package is installed (e.g., `ath10k-firmware-qca988x`)
- **Driver fails to load**: Some devices require specific firmware versions or additional packages
- **Wrong firmware**: Some devices (e.g., QCA9887) need different firmware than similar models (e.g., QCA988X). Check `lspci` output for exact device ID.