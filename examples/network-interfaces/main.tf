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
  ula_prefix     = "fd00:cafe::/48"
  packet_steering = true
}

resource "openwrt_network_device" "br_lan" {
  name  = "br-lan"
  type  = "bridge"
  ports = "eth0 eth1"
}

resource "openwrt_network_device" "br_guest" {
  name  = "br-guest"
  type  = "bridge"
}

resource "openwrt_network_bridge_vlan" "guest_vlan10" {
  device = "br-guest"
  vlan   = 10
  ports  = {
    "eth2" = "t"
  }
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
  metric  = 100
  auto    = "1"

  lifecycle {
    ignore_changes = [ipaddr]
  }
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
  device  = "br-guest.10"
  ipaddr  = ["10.10.10.1/24"]
  ip6addr = ["fd00:cafe::1/64"]
  auto    = "1"
}

resource "openwrt_network_interface" "wireguard" {
  name    = "wg0"
  proto   = "wireguard"
  ipaddr  = ["10.20.0.1/24"]
  ip6addr = ["fd00:wg::1/64"]
  auto    = "1"
}

resource "openwrt_network_wireguard" "peer_nordvpn" {
  name         = "peer1_wg0"
  description  = "Norway VPN Server"
  public_key   = "server_public_key_here"
  endpoint_host = "no123.nordvpn.com"
  endpoint_port = 51820
  persistent_keepalive = 25
  allowed_ips   = "0.0.0.0/0"
}

output "lan_ip" {
  description = "LAN interface IP"
  value       = openwrt_network_interface.lan.ipaddr
}

output "guest_ip" {
  description = "Guest network IP"
  value       = openwrt_network_interface.guest.ipaddr
}

output "wireguard_ip" {
  description = "WireGuard interface IP"
  value       = openwrt_network_interface.wireguard.ipaddr
}
