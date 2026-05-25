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

resource "openwrt_network_globals" "main" {
  ula_prefix      = "fd00:cafe::/48"
  packet_steering = true
}

resource "openwrt_network_device" "br_lan" {
  name  = "br-lan"
  type  = "bridge"
  ports = ["eth0", "eth1"]
}

resource "openwrt_network_device" "br_guest" {
  name  = "br-guest"
  type  = "bridge"
  ports = ["eth2"]
}

resource "openwrt_network_interface" "loopback" {
  name    = "loopback"
  proto   = "static"
  device  = "lo"
  ipaddr  = ["127.0.0.1/8"]
}

resource "openwrt_network_interface" "lan" {
  name    = "lan"
  proto   = "static"
  device  = "br-lan"
  ipaddr  = ["192.168.1.1/24"]
  ip6addr = ["fd00:lan::1/64"]
  metric  = 100
  auto    = "1"
}

resource "openwrt_network_interface" "wan" {
  name   = "wan"
  proto  = "dhcp"
  device = "eth0"
  auto   = "1"
}

resource "openwrt_network_interface" "guest" {
  name    = "guest"
  proto   = "static"
  device  = "br-guest"
  ipaddr  = ["10.10.10.1/24"]
  ip6addr = ["fd00:guest::1/64"]
  auto    = "1"
}

resource "openwrt_network_bridge_vlan" "guest_vlan10" {
  device = "br-guest"
  vlan   = 10
  ports  = {
    "eth2" = "t"
  }
}

resource "openwrt_firewall_defaults" "main" {
  input             = "REJECT"
  output            = "ACCEPT"
  forward           = "REJECT"
  synflood_protect  = true
  synflood_rate     = "25/s"
  drop_invalid      = true
  auto_helper       = true
}

resource "openwrt_firewall_zone" "lan" {
  name     = "lan"
  input   = "ACCEPT"
  output  = "ACCEPT"
  forward = "REJECT"
  masq    = false
  network = ["lan"]
}

resource "openwrt_firewall_zone" "wan" {
  name     = "wan"
  input   = "DROP"
  output  = "ACCEPT"
  forward = "REJECT"
  masq    = true
  mtu_fix = true
  network = ["wan"]
}

resource "openwrt_firewall_zone" "guest" {
  name     = "guest"
  input   = "REJECT"
  output  = "ACCEPT"
  forward = "REJECT"
  masq    = true
  network = ["guest"]
}

resource "openwrt_firewall_forwarding" "lan_to_wan" {
  src  = "lan"
  dest = "wan"
}

resource "openwrt_firewall_forwarding" "guest_to_wan" {
  src  = "guest"
  dest = "wan"
}

resource "openwrt_firewall_rule" "allow_dhcp_lan" {
  name       = "Allow-DHCP-LAN"
  src        = "lan"
  proto      = "udp"
  dest_port  = "67"
  target     = "ACCEPT"
  family     = "ipv4"
}

resource "openwrt_firewall_rule" "allow_dns_lan" {
  name       = "Allow-DNS-LAN"
  src        = "lan"
  proto      = "tcpudp"
  dest_port  = "53"
  target     = "ACCEPT"
}

resource "openwrt_firewall_rule" "allow_dns_guest" {
  name       = "Allow-DNS-Guest"
  src        = "guest"
  proto      = "tcpudp"
  dest_port  = "53"
  target     = "ACCEPT"
}

resource "openwrt_firewall_rule" "allow_dhcp_guest" {
  name       = "Allow-DHCP-Guest"
  src        = "guest"
  proto      = "udp"
  dest_port  = "67"
  target     = "ACCEPT"
  family     = "ipv4"
}

resource "openwrt_firewall_rule" "allow_ping_lan" {
  name   = "Allow-Ping-LAN"
  src    = "lan"
  proto  = "icmp"
  target = "ACCEPT"
  family = "ipv4"
}

resource "openwrt_firewall_rule" "allow_ping_wan" {
  name      = "Allow-Ping-WAN"
  src       = "wan"
  proto     = "icmp"
  icmp_type = "echo-request"
  target    = "ACCEPT"
  family    = "ipv4"
}

resource "openwrt_firewall_rule" "allow_ssh_lan" {
  name       = "Allow-SSH-LAN"
  src        = "lan"
  proto      = "tcp"
  dest_port  = "22"
  target     = "ACCEPT"
  family     = "ipv4"
}

resource "openwrt_firewall_rule" "allow_http_lan" {
  name       = "Allow-HTTP-LAN"
  src        = "lan"
  proto      = "tcp"
  dest_port  = "80"
  target     = "ACCEPT"
  family     = "ipv4"
}

resource "openwrt_firewall_rule" "allow_https_lan" {
  name       = "Allow-HTTPS-LAN"
  src        = "lan"
  proto      = "tcp"
  dest_port  = "443"
  target     = "ACCEPT"
  family     = "ipv4"
}

resource "openwrt_firewall_rule" "block_guest_to_lan" {
  name   = "Block-Guest-to-LAN"
  src    = "guest"
  dest   = "lan"
  target = "DROP"
  family = "ipv4"
}

resource "openwrt_firewall_rule" "allow_dhcpv6_wan" {
  name       = "Allow-DHCPv6-WAN"
  src        = "wan"
  proto      = "udp"
  dest_port  = "546"
  target     = "ACCEPT"
  family     = "ipv6"
}

resource "openwrt_firewall_rule" "allow_icmpv6_wan" {
  name       = "Allow-ICMPv6-WAN"
  src        = "wan"
  proto      = "icmp"
  target     = "ACCEPT"
  family     = "ipv6"
  icmp_type  = "echo-request echo-reply destination-unreachable packet-too-big time-exceeded"
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
  resolvfile        = "/tmp/resolv.conf.d/resolv.conf.auto"
  localservice     = true
  server           = "1.1.1.1 8.8.8.8"
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
  ra_flags   = "managed-config other-config"
}

resource "openwrt_dhcp_pool" "guest" {
  name       = "guest"
  interface  = "guest"
  start      = 50
  limit      = 50
  leasetime  = "6h"
  dhcpv4     = "server"
  ra         = "disabled"
}

resource "openwrt_dhcp_host" "nas" {
  name = "nas"
  ip   = "192.168.1.10"
  mac  = "00:11:22:33:44:55"
}

resource "openwrt_dhcp_host" "printer" {
  name = "printer"
  ip   = "192.168.1.20"
  mac  = "AA:BB:CC:DD:EE:FF"
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

resource "openwrt_wireless_iface" "home_2g" {
  name        = "wifinet0"
  device      = "radio0"
  mode        = "ap"
  ssid        = "HomeNet"
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

resource "openwrt_wireless_iface" "guest_2g" {
  name        = "wifinet2"
  device      = "radio0"
  mode        = "ap"
  ssid        = "GuestNet"
  encryption  = "psk2"
  key         = "guest_password"
  network     = ["guest"]
  hidden      = false
  isolate     = true
}

resource "openwrt_wireless_iface" "guest_5g" {
  name        = "wifinet3"
  device      = "radio1"
  mode        = "ap"
  ssid        = "GuestNet-5GHz"
  encryption  = "psk2"
  key         = "guest_password"
  network     = ["guest"]
  hidden      = false
  isolate     = true
}

resource "openwrt_system" "main" {
  hostname     = "openwrt-router"
  ttylogin     = true
  urandom_seed = true
  zonename     = "America/New_York"
  log_proto    = "udp"
  conloglevel  = "7"
}

resource "openwrt_system_ntp" "main" {
  name    = "ntp"
  enabled = true
  server  = "0.openwrt.pool.ntp.org 1.openwrt.pool.ntp.org 2.openwrt.pool.ntp.org"
}

resource "openwrt_dropbear" "main" {
  name                = "main"
  password_auth       = true
  port                = 22
  root_password_auth  = false
  root_login          = true
}

output "lan_interface_ip" {
  description = "LAN interface IP address"
  value       = openwrt_network_interface.lan.ipaddr
}

output "guest_interface_ip" {
  description = "Guest interface IP address"
  value       = openwrt_network_interface.guest.ipaddr
}

output "lan_zone_id" {
  description = "LAN firewall zone ID"
  value       = openwrt_firewall_zone.lan.id
}

output "wan_zone_id" {
  description = "WAN firewall zone ID"
  value       = openwrt_firewall_zone.wan.id
}
