terraform {
  required_providers {
    openwrt = {
      source  = "bucksbunni/openwrt"
      version = "~> 0.1.0"
    }
  }
}

provider "openwrt" {
  host     = "http://192.168.1.1"
  username = "root"
  password = "your_password"
  insecure = true
}

resource "openwrt_system" "main" {
  hostname      = "openwrt-router"
  ttylogin      = true
  log_size      = 128
  urandom_seed  = true
  zonename      = "America/New_York"
  log_proto     = "udp"
  conloglevel   = "7"
  cronloglevel  = "8"
}

resource "openwrt_system_ntp" "main" {
  name    = "ntp"
  enabled = true
  server  = "0.openwrt.pool.ntp.org 1.openwrt.pool.ntp.org 2.openwrt.pool.ntp.org 3.openwrt.pool.ntp.org"
  port    = 123
}

resource "openwrt_system_led" "wan_led" {
  name     = "wan"
  sysfs    = "apu:green:3"
  trigger = "netdev"
  mode    = "link tx rx"
  dev     = "wan"
  default = true
}

resource "openwrt_system_led" "wifi_2g_led" {
  name     = "wifi-2g"
  sysfs    = "apu:green:2"
  trigger = "netdev"
  mode    = "tx rx"
  dev     = "wlan0"
  default = true
}

resource "openwrt_system_led" "wifi_5g_led" {
  name     = "wifi-5g"
  sysfs    = "apu:green:1"
  trigger = "netdev"
  mode    = "tx rx"
  dev     = "wlan1"
  default = true
}

resource "openwrt_system_led" "heartbeat" {
  name     = "status-heartbeat"
  sysfs    = "apu:green:1"
  trigger = "heartbeat"
  default = true
}

resource "openwrt_system_led" "usb_led" {
  name     = "usb"
  sysfs    = "apu:green:2"
  trigger = "netdev"
  mode    = "link"
  dev     = "usb0"
  default = false
}

resource "openwrt_dropbear" "main" {
  name              = "main"
  password_auth     = true
  port              = 22
  root_password_auth = false
  root_login        = true
}

resource "openwrt_dropbear" "backup" {
  name          = "backup"
  password_auth = true
  port          = 2222
  root_login    = false
}

output "system_id" {
  description = "System settings ID"
  value       = openwrt_system.main.id
}

output "hostname" {
  description = "Router hostname"
  value       = openwrt_system.main.hostname
}

output "ntp_id" {
  description = "NTP configuration ID"
  value       = openwrt_system_ntp.main.id
}
