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

resource "openwrt_dhcp_dnsmasq" "main" {
  domainneeded     = true
  localise_queries = true
  rebind_protection = true
  rebind_localhost  = true
  local            = "/lan/"
  domain           = "lan"
  expand_hosts     = true
  cachesize        = 1000
  authoritative    = true
  readethers       = true
  leasefile        = "/tmp/dhcp.leases"
  resolvfile       = "/tmp/resolv.conf.d/resolv.conf.auto"
  localservice     = true
  ednspacket_max   = 1232
  server           = "1.1.1.1 8.8.8.8"
  rebind_domain     = "local FDE1:B7F5:46F7::/48"
}

resource "openwrt_dhcp_odhcpd" "main" {
  maindhcp     = true
  leasefile    = "/tmp/dhcp.leases"
  leasetrigger = "/usr/sbin/odhcpd-update"
  loglevel     = 4
}

resource "openwrt_dhcp_pool" "lan" {
  name       = "lan"
  interface  = "lan"
  start      = 100
  limit      = 150
  leasetime  = "12h"
  dhcpv4     = "server"
  ra         = "server"
  dhcpv6     = "server"
  ra_flags   = ["managed-config", "other-config"]
}

resource "openwrt_dhcp_pool" "guest" {
  name       = "guest"
  interface  = "guest"
  start      = 50
  limit      = 50
  leasetime  = "6h"
  dhcpv4     = "server"
  ra         = "disabled"
  dhcpv6     = "disabled"
}

resource "openwrt_dhcp_pool" "iot" {
  name       = "iot"
  interface  = "iot"
  start      = 10
  limit      = 50
  leasetime  = "24h"
  dhcpv4     = "server"
  ra         = "disabled"
}

resource "openwrt_dhcp_pool" "guest_v6" {
  name       = "guest_v6"
  interface  = "guest"
  start      = 100
  limit      = 50
  leasetime  = "6h"
  dhcpv4     = "disabled"
  ra         = "server"
  dhcpv6     = "server"
  ra_flags   = ["managed-config"]
}

resource "openwrt_dhcp_host" "server_static" {
  name = "fileserver"
  ip   = "192.168.1.10"
  mac  = ["00:11:22:33:44:55"]
}

resource "openwrt_dhcp_host" "printer" {
  name      = "printer"
  ip        = "192.168.1.20"
  mac       = ["AA:BB:CC:DD:EE:FF"]
  leasetime = "infinite"
}

resource "openwrt_dhcp_host" "nas" {
  name = "nas"
  ip   = "192.168.1.30"
  mac  = ["11:22:33:44:55:66"]
}

resource "openwrt_dhcp_host" "dev_machine" {
  name      = "dev-machine"
  ip        = "10.10.10.50"
  mac       = ["DE:AD:BE:EF:CA:FE"]
  leasetime = "24h"
}

resource "openwrt_dhcp_host" "iot_camera" {
  name = "camera-1"
  ip   = "192.168.2.100"
  mac  = ["C0:FF:EE:C0:FF:EE"]
}

output "lan_pool_id" {
  description = "LAN DHCP pool ID"
  value       = openwrt_dhcp_pool.lan.id
}

output "guest_pool_id" {
  description = "Guest DHCP pool ID"
  value       = openwrt_dhcp_pool.guest.id
}

output "dnsmasq_id" {
  description = "DNS/DHCP server ID"
  value       = openwrt_dhcp_dnsmasq.main.id
}
