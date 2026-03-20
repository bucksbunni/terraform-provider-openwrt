# Complete Router Configuration

This directory contains a comprehensive Terraform configuration that sets up a complete OpenWrt router with all common features.

## Files

- `main.tf` - Full router configuration (network, firewall, DHCP, wireless, system)
- `provider.tf` - Reusable provider configuration with variables
- `README.md` - This file

## What This Creates

### Network

- LAN bridge (br-lan) with eth0, eth1
- Guest bridge (br-guest) with VLAN 10
- Static LAN interface (192.168.1.1/24)
- DHCP WAN interface
- Guest interface (10.10.10.1/24)
- Network globals with ULA prefix and packet steering

### Firewall

- LAN, WAN, Guest, and VPN zones
- Zone-to-zone forwarding rules
- Default policies with SYN flood protection
- Allow rules for DHCP, DNS, SSH, HTTP(S) from LAN
- ICMP and DHCPv6 rules for WAN
- Inter-zone blocking rules

### DHCP/DNS

- dnsmasq with rebind protection and upstream DNS
- DHCP pools for LAN (100-249) and Guest (50-99)
- IPv6 DHCP with RA settings
- Static host reservations

### Wireless

- Dual-band (2.4GHz + 5GHz) WiFi radios
- Home SSID on both bands with WPA2
- Guest SSID on both bands with client isolation
- WPA3 on 5GHz band

### System

- Hostname, timezone, logging
- NTP client with pool servers
- Status LEDs for WAN and WiFi
- Dropbear SSH on port 22

## Usage

```bash
terraform init
terraform plan
terraform apply
```

## Important Notes

- Review and update all passwords before applying
- Check channel settings comply with local regulations
- Adjust IP ranges to match your network plan
- The complete configuration uses `terraform apply` with auto-approve for demonstration
