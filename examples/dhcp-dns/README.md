# DHCP and DNS Management Examples

This directory contains Terraform configurations for managing DHCP pools, DNS settings, and static DHCP reservations on OpenWrt.

## Files

- `main.tf` - Complete DHCP and DNS configuration with provider setup
- `provider.tf` - Reusable provider configuration with variables
- `README.md` - This file

## Resources Created

### DHCP Pools

- `openwrt_dhcp_pool` for LAN, Guest, and IoT networks
- Each pool configured with appropriate IP ranges, lease times, and IPv6 settings

### DNS/DHCP Server (dnsmasq)

- `openwrt_dhcp_dnsmasq` for global DNS and DHCP settings
- Rebind protection, local domain, upstream DNS servers

### DHCPv6 Server (odhcpd)

- `openwrt_dhcp_odhcpd` for IPv6 DHCP server configuration

### Static Host Reservations

- `openwrt_dhcp_host` for devices with fixed IP addresses
- Maps MAC addresses to reserved IPs

## Usage

```bash
terraform init
terraform plan
terraform apply
```

## Variables

Configure via `provider.tf` or environment variables:
- `openwrt_host` - LuCI URL (default: http://192.168.1.1)
- `openwrt_username` - SSH username (default: root)
- `openwrt_password` - SSH password
- `insecure` - Skip TLS verification (default: true)
