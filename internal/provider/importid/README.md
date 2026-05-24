# Import ID Parser Package

This package provides utilities for parsing and formatting Terraform resource import IDs for OpenWrt UCI configuration sections.

## Overview

When importing existing OpenWrt configuration into Terraform, the import ID encodes both the UCI config file and section name. This package provides consistent parsing and formatting of these IDs.

## ID Format

All import IDs follow the pattern: `<config>/<section_name>`

Examples:
- `network/lan` - Network interface named "lan"
- `firewall/wan` - Firewall zone named "wan"
- `dhcp/lan` - DHCP pool for interface "lan"
- `wireless/radio0` - Wireless device "radio0"

## Parser Functions

Each resource type has dedicated parser and formatter functions:

### Network Resources

| Function | Description |
|----------|-------------|
| `ParseNetworkInterfaceID(id)` | Parse network interface ID, returns name |
| `FormatNetworkInterfaceID(name)` | Format name into network interface ID |

### DHCP Resources

| Function | Description |
|----------|-------------|
| `ParseDHCPPoolID(id)` | Parse DHCP pool ID, returns pool name |
| `FormatDHCPPoolID(name)` | Format name into DHCP pool ID |
| `ParseDHCPHostID(id)` | Parse DHCP host ID, returns host name |
| `FormatDHCPHostID(name)` | Format name into DHCP host ID |

### Firewall Resources

| Function | Description |
|----------|-------------|
| `ParseFirewallZoneID(id)` | Parse firewall zone ID, returns zone name |
| `FormatFirewallZoneID(name)` | Format name into firewall zone ID |
| `ParseFirewallRuleID(id)` | Parse firewall rule ID, returns rule name |
| `FormatFirewallRuleID(name)` | Format name into firewall rule ID |
| `ParseFirewallForwardingID(id)` | Parse forwarding ID, returns src and dest |
| `FormatFirewallForwardingID(src, dest)` | Format src and dest into forwarding ID |

### Other Resources

| Function | Description |
|----------|-------------|
| `ParseWirelessInterfaceID(id)` | Parse wireless interface ID |
| `FormatWirelessInterfaceID(name)` | Format name into wireless interface ID |
| `ParseDropbearID(id)` | Parse Dropbear SSH instance ID |
| `FormatDropbearID(name)` | Format name into Dropbear ID |
| `ParseSystemLEDID(id)` | Parse system LED ID |
| `FormatSystemLEDID(name)` | Format name into system LED ID |

## Core Functions

| Function | Description |
|----------|-------------|
| `ParseUCISectionID(id)` | Parse generic UCI section ID, returns config and name |
| `FormatUCISectionID(config, name)` | Format config and name into UCI section ID |

## Usage Example

```go
// In resource ImportState handler
func (r *networkInterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

    name, err := importid.ParseNetworkInterfaceID(req.ID)
    if err != nil {
        resp.Diagnostics.AddError("Invalid import ID", err.Error())
        return
    }

    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
}
```

## Terraform Import Command

```bash
# Import a network interface
terraform import openwrt_network_interface.lan network/lan

# Import a firewall zone
terraform import openwrt_firewall_zone.wan firewall/wan

# Import a DHCP pool
terraform import openwrt_dhcp_pool.lan dhcp/lan
```

## Testing

Parser tests verify:
- Valid ID parsing returns correct components
- Invalid ID formats return appropriate errors
- ID formatting produces correct format

Run tests with:
```bash
go test ./internal/provider/importid/...
```
