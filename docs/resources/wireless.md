---
layout: ""
page_title: "Wireless Resources"
description: |-
  Wireless resource types for the OpenWrt provider.
---

# Wireless Resources

For detailed setup instructions, including driver and firmware installation, see the [Wireless Setup Guide](../guides/wireless-setup.md).

-> **Important:** After installing kernel module packages (e.g., `kmod-ath10k`), you must either:
> 1. Use `openwrt_sys_modprobe` to load the driver module, or
> 2. Use `openwrt_sys_reboot` to reboot the device so the module loads automatically
>
> The wireless radio will not appear until the kernel module is loaded. Use `openwrt_sys_wireless_info` data source to verify the radio is detected.

## openwrt_wireless_device

Manages wireless radio devices.

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID: wireless/<device_name> |
| `name` | String | Yes | Radio name (e.g., 'radio0', 'radio1') |
| `type` | String | No | Driver type (e.g., 'mac80211', 'broadcom') |
| `path` | String | No | Hardware path |
| `band` | String | No | Radio band: '2g', '5g', '6g' |
| `channel` | Int64 | No | Wireless channel number |
| `htmode` | String | No | Channel width (e.g., 'VHT80', 'HT40', 'HE80') |
| `hwmode` | String | No | Hardware mode: '11a', '11g', '11n', '11ac', '11ax'. Auto-detected if not set. |
| `country` | String | No | Country code |
| `disabled` | Bool | No | Disable the radio |

### Mode Selection

The wireless mode is controlled by `hwmode`:
- `11g` - 2.4GHz, 802.11b/g only (Legacy)
- `11n` - 2.4GHz, 802.11n
- `11a` - 5GHz, 802.11a only (Legacy)
- `11ac` - 5GHz, 802.11ac
- `11ax` - 6GHz, 802.11ax

If not set, the mode is auto-detected based on `band` and `htmode`:
- `band="2g"` + `htmode="HT20"` or `htmode="HT40"` = 802.11n
- `band="5g"` + `htmode="HT20"` or `htmode="HT40"` = 802.11n
- `band="5g"` + `htmode="VHT80"` = 802.11ac

### Example

```hcl
# 802.11ac (5GHz)
resource "openwrt_wireless_device" "radio0_5g" {
  name     = "radio0"
  type     = "mac80211"
  path     = "pci0000:00/0000:00:02.5/0000:04:00.0"
  band     = "5g"
  channel  = 44
  htmode   = "VHT80"
  hwmode   = "11ac"
  country  = "DE"
  disabled = false
}

# 802.11n (2.4GHz)
resource "openwrt_wireless_device" "radio1_2g" {
  name     = "radio1"
  type     = "mac80211"
  band     = "2g"
  channel  = 6
  htmode   = "HT40"
  hwmode   = "11n"
  disabled = false
}
```

---

## openwrt_wireless_iface

Manages wireless interfaces (SSIDs).

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | String | Computed | Internal ID |
| `name` | String | Yes | Interface name (e.g., 'wifinet0') |
| `device` | String | Yes | Radio device (e.g., 'radio0') |
| `mode` | String | No | Mode: 'ap', 'sta', 'mesh', 'adhoc' |
| `ssid` | String | No | SSID name |
| `encryption` | String | No | Encryption type: 'psk2', 'psk', 'wep', 'none' |
| `key` | String | No | Encryption key/password |
| `network` | String | No | Network to attach |
| `disabled` | Bool | No | Disable the interface |
| `hidden` | Bool | No | Hide SSID |
| `macfilter` | String | No | MAC filter: 'disable', 'allow', 'deny', 'radius' |
| `maclist` | String | No | MAC address list |
| `isolate` | Bool | No | Client isolation |

### Example

```hcl
# Access Point
resource "openwrt_wireless_iface" "main" {
  name        = "wifinet0"
  device      = "radio0"
  mode        = "ap"
  ssid        = "MyNetwork"
  encryption  = "psk2"
  key         = "secretpassword"
  network     = "lan"
  hidden      = false
  isolate     = false
}

# Guest Network
resource "openwrt_wireless_iface" "guest" {
  name        = "wifinet1"
  device      = "radio0"
  mode        = "ap"
  ssid        = "GuestNetwork"
  encryption  = "psk2"
  key         = "guestpassword"
  network     = "guest"
  isolate     = true
}
```

---

## Import

```bash
terraform import openwrt_wireless_device.radio0 radio0
terraform import openwrt_wireless_iface.main wifinet0
```

## Full Example

See [examples/wireless](../../examples/wireless/) for a complete Terraform configuration with multiple SSIDs.