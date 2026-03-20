# Firewall Setup Example

This example demonstrates creating a comprehensive firewall configuration with zones, rules, and forwarding on an OpenWrt router.

## Overview

Creates:
- Firewall defaults (REJECT policy with SYN flood protection)
- LAN zone (trusted, full access)
- WAN zone (untrusted, NAT/masquerade enabled)
- Guest zone (isolated, limited access)
- VPN zones (WireGuard tunnels)
- Common firewall rules (DHCP, DNS, ping)
- Custom rules (HTTP, SSH from LAN only)
- Zone forwarding rules

## Files

- `main.tf` - Complete firewall configuration
- `provider.tf` - Provider configuration

## Usage

```bash
terraform init
terraform plan
terraform apply
```

## Security Zones

| Zone | Input | Output | Forward | Masquerade |
|------|-------|--------|---------|------------|
| lan  | ACCEPT | ACCEPT | REJECT | No |
| wan  | DROP | ACCEPT | REJECT | Yes |
| guest | REJECT | ACCEPT | REJECT | Yes |
| vpn | DROP | REJECT | REJECT | No |

## Forwarding Rules

- LAN → WAN (allow)
- LAN → VPN (allow)
- Guest → WAN (allow)
- VPN → LAN (allow for specific services)

## Common Ports Opened

| Service | Protocol | Port | Source |
|---------|----------|------|--------|
| DHCP | UDP | 67-68 | Any |
| DNS | TCP/UDP | 53 | LAN, Guest |
| Ping | ICMP | - | LAN |
| SSH | TCP | 22 | LAN only |
