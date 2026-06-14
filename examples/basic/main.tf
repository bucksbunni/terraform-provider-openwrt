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

data "openwrt_sys_hostname" "router" {}

data "openwrt_sys_uptime" "status" {}

output "hostname" {
  description = "Router hostname"
  value       = data.openwrt_sys_hostname.router.hostname
}

output "uptime_seconds" {
  description = "Router uptime in seconds"
  value       = data.openwrt_sys_uptime.status.uptime
}

output "network_interfaces" {
  description = "Available network interfaces"
  value       = data.openwrt_sys_net_devices.main.devices
}
