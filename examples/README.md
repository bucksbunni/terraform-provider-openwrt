# OpenWrt Terraform Provider Examples

This directory contains example Terraform configurations demonstrating the use of the OpenWrt provider.

## Examples

### [basic](./basic/)

Simple example showing data source usage. Reads system hostname, uptime, and network device information from OpenWrt.

### [network-interfaces](./network-interfaces/)

Network configuration including:
- Bridge devices (br-lan, br-guest)
- Static and DHCP interfaces (LAN, WAN, Guest)
- Bridge VLANs
- WireGuard VPN peers

### [firewall-setup](./firewall-setup/)

Firewall configuration with:
- Multiple zones (LAN, WAN, Guest, VPN)
- Zone-to-zone forwarding
- Rules for DHCP, DNS, ICMP, SSH, HTTP(S)
- IPv6 firewall rules
- SYN flood protection defaults

### [dhcp-dns](./dhcp-dns/)

DHCP and DNS management:
- dnsmasq configuration with rebind protection
- DHCP pools for multiple networks (LAN, Guest, IoT)
- IPv6 DHCP with odhcpd
- Static host reservations by MAC address

### [wireless](./wireless/)

WiFi configuration:
- Dual-band radio setup (2.4GHz, 5GHz, 6GHz)
- Multiple SSIDs per band (Home, Guest, IoT, Staff)
- WPA2/WPA3 encryption
- MAC filtering and client isolation

### [system](./system/)

System settings management:
- Hostname, timezone, logging configuration
- NTP client with pool servers
- LED triggers (netdev, heartbeat, timer)
- Dropbear SSH server instances

### [complete-router](./complete-router/)

Full router configuration combining all resources:
- Complete network setup with bridges and VLANs
- Multi-zone firewall with comprehensive rules
- DHCP/DNS with static reservations
- Dual-band wireless with multiple SSIDs
- System configuration and SSH access

## Usage

Each example directory is self-contained. To use an example:

```bash
cd examples/<example-name>
terraform init
terraform plan
terraform apply
```

## Provider Configuration

Examples use the `bucksbunni/openwrt` provider from the Terraform Registry. Configure the provider with:

- `host` - LuCI URL (e.g., `http://192.168.1.1`)
- `username` - SSH username (default: `root`)
- `password` - SSH password
- `insecure` - Skip TLS verification (default: `true` for self-signed certs)

For production use, consider:
- Using environment variables for credentials
- Using Terraform workspace-specific configurations
- Setting up proper TLS certificates to disable `insecure`
