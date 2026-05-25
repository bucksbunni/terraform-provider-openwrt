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
  password = "your_password"
  insecure = true
}

resource "openwrt_wireless_device" "radio0" {
  name     = "radio0"
  type     = "mac80211"
  band     = "2g"
  channel  = 6
  htmode   = "HT40"
  country  = "US"
  disabled = false
}

resource "openwrt_wireless_device" "radio1" {
  name     = "radio1"
  type     = "mac80211"
  band     = "5g"
  channel  = 36
  htmode   = "VHT80"
  country  = "US"
  disabled = false
}

resource "openwrt_wireless_device" "radio2" {
  name     = "radio2"
  type     = "mac80211"
  band     = "6g"
  channel  = 1
  htmode   = "HE80"
  country  = "US"
  disabled = false
}

resource "openwrt_wireless_iface" "home_2g" {
  name        = "wifinet0"
  device      = "radio0"
  mode        = "ap"
  ssid        = "HomeNet-2.4GHz"
  encryption  = "psk2+ccmp"
  key         = "your_wpa2_password"
  network     = ["lan"]
  hidden      = false
  isolate     = false
}

resource "openwrt_wireless_iface" "home_5g" {
  name        = "wifinet1"
  device      = "radio1"
  mode        = "ap"
  ssid        = "HomeNet-5GHz"
  encryption  = "psk2+ccmp"
  key         = "your_wpa2_password"
  network     = ["lan"]
  hidden      = false
  isolate     = false
}

resource "openwrt_wireless_iface" "home_6g" {
  name        = "wifinet2"
  device      = "radio2"
  mode        = "ap"
  ssid        = "HomeNet-6GHz"
  encryption  = "sae"
  key         = "your_wpa3_password"
  network     = ["lan"]
  hidden      = false
  isolate     = false
}

resource "openwrt_wireless_iface" "guest_2g" {
  name        = "wifinet3"
  device      = "radio0"
  mode        = "ap"
  ssid        = "GuestNet"
  encryption  = "psk2"
  key         = "guest_password_2024"
  network     = ["guest"]
  hidden      = false
  isolate     = true
}

resource "openwrt_wireless_iface" "guest_5g" {
  name        = "wifinet4"
  device      = "radio1"
  mode        = "ap"
  ssid        = "GuestNet-5GHz"
  encryption  = "psk2"
  key         = "guest_password_2024"
  network     = ["guest"]
  hidden      = false
  isolate     = true
}

resource "openwrt_wireless_iface" "iot_2g" {
  name        = "wifinet5"
  device      = "radio0"
  mode        = "ap"
  ssid        = "IoT-Network"
  encryption  = "psk2"
  key         = "iot_secure_pass"
  network     = ["iot"]
  hidden      = false
  isolate     = true
  macfilter   = "allow"
  maclist     = ["AA:BB:CC:DD:EE:01", "AA:BB:CC:DD:EE:02", "AA:BB:CC:DD:EE:03"]
}

resource "openwrt_wireless_iface" "staff_5g" {
  name        = "wifinet6"
  device      = "radio1"
  mode        = "ap"
  ssid        = "Staff-Network"
  encryption  = "psk2+ccmp"
  key         = "staff_password_2024"
  network     = ["staff"]
  hidden      = false
  isolate     = false
  macfilter   = "allow"
  maclist     = ["11:22:33:44:55:66", "77:88:99:AA:BB:CC"]
}

output "radio0_id" {
  description = "Radio0 device ID"
  value       = openwrt_wireless_device.radio0.id
}

output "radio1_id" {
  description = "Radio1 device ID"
  value       = openwrt_wireless_device.radio1.id
}

output "home_ssid_2g_id" {
  description = "Home 2.4GHz SSID ID"
  value       = openwrt_wireless_iface.home_2g.id
}
