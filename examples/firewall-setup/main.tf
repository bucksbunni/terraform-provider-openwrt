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

resource "openwrt_firewall_zone" "vpn" {
  name     = "vpn"
  input   = "DROP"
  output  = "REJECT"
  forward = "REJECT"
  masq    = false
  network = ["wg0"]
}

resource "openwrt_firewall_forwarding" "lan_to_wan" {
  src  = "lan"
  dest = "wan"
}

resource "openwrt_firewall_forwarding" "lan_to_vpn" {
  src  = "lan"
  dest = "vpn"
}

resource "openwrt_firewall_forwarding" "guest_to_wan" {
  src  = "guest"
  dest = "wan"
}

resource "openwrt_firewall_forwarding" "vpn_to_lan" {
  src  = "vpn"
  dest = "lan"
}

resource "openwrt_firewall_rule" "allow_dhcp_lan" {
  name       = "Allow-DHCP-LAN"
  src        = "lan"
  proto      = "udp"
  dest_port  = "67"
  target     = "ACCEPT"
  family     = "ipv4"
}

resource "openwrt_firewall_rule" "allow_dhcp_guest" {
  name       = "Allow-DHCP-Guest"
  src        = "guest"
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

resource "openwrt_firewall_rule" "allow_ping_lan" {
  name      = "Allow-Ping-LAN"
  src       = "lan"
  proto     = "icmp"
  target    = "ACCEPT"
  family    = "ipv4"
}

resource "openwrt_firewall_rule" "allow_ping_wan" {
  name      = "Allow-Ping-WAN"
  src       = "wan"
  proto     = "icmp"
  icmp_type = ["echo-request"]
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
  name      = "Block-Guest-to-LAN"
  src       = "guest"
  dest      = "lan"
  target    = "DROP"
  family    = "ipv4"
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
  icmp_type  = ["echo-request", "echo-reply", "destination-unreachable", "packet-too-big", "time-exceeded"]
}

output "lan_zone_id" {
  description = "LAN firewall zone ID"
  value       = openwrt_firewall_zone.lan.id
}

output "wan_zone_id" {
  description = "WAN firewall zone ID"
  value       = openwrt_firewall_zone.wan.id
}

output "guest_zone_id" {
  description = "Guest firewall zone ID"
  value       = openwrt_firewall_zone.guest.id
}
