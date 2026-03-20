# Network Interfaces Example

This example demonstrates creating network interfaces, bridges, VLANs, and WireGuard VPN on an OpenWrt router.

## Overview

Creates:
- LAN interface (bridge with static IP)
- WAN interface (DHCP client)
- Guest network (isolated VLAN)
- Bridge device with ports
- Bridge VLAN assignments
- Network globals (ULA prefix, packet steering)
- WireGuard VPN peer

## Files

- `main.tf` - Complete network configuration
- `provider.tf` - Provider configuration

## Usage

```bash
terraform init
terraform plan
terraform apply
```

## Network Diagram

```
          +----------------+
          |  OpenWrt Router |
          +----------------+
                 |
    +------------+------------+
    |                         |
  br-lan                   eth0 (WAN)
    |                         |
  LAN (192.168.1.0/24)   DHCP from ISP
    |
+---+----+              +----+-----+
|         |              |          |
wlan0   wlan1         br-guest  WireGuard
(Home)  (Guest)        (VLAN 10)   (VPN)
```

## Security Considerations

- Guest network is isolated from LAN
- WireGuard uses persistent keepalive for NAT traversal
- ULA prefix provides IPv6 isolation
