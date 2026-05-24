# Converters Package

This package provides data model converters for translating between Terraform types and OpenWrt UCI configuration values.

## Overview

The converters handle the translation between:

1. **Terraform Framework Types** (`*types.String`, `*types.Bool`, etc.) - Used in Terraform resource schemas
2. **UCI Options Maps** (`map[string]interface{}`) - Used when calling the LuCI RPC API
3. **Converter Options Structs** - Intermediate Go structs with pointer fields for precise value tracking

## Why Pointer Fields?

The converter option structs use pointer fields (`*string`, `*bool`, `*int64`) to distinguish between:

- **Nil** (not set/not specified)
- **Zero value** (explicitly set to empty/default)

This distinction is important for Terraform to properly detect when a user has explicitly set a value versus when it should use a computed default.

## Available Converters

### Firewall

- `FirewallZoneOptions` - Firewall zone settings
- `FirewallRuleOptions` - Firewall rule settings
- `FirewallForwardingOptions` - Zone-to-zone forwarding

### Network

- `NetworkInterfaceOptions` - Network interface configuration
- `WirelessInterfaceOptions` - Wireless interface (WiFi) configuration

### DHCP

- `DHCPPoolOptions` - DHCP pool configuration
- `DNSMasqOptions` - DNS/DHCP (dnsmasq) settings
- `DHCPHostOptions` - Static DHCP host reservation

### System

- `SystemLEDOptions` - LED configuration

## Common Utilities (`uci.go`)

- `BoolToUCI(bool) string` - Convert bool to UCI "1"/"0"
- `ParseUCIBool(interface{}) (bool, bool)` - Parse UCI bool with null detection
- `ParseUCIInt(interface{}) (int64, bool)` - Parse UCI integer with null detection
- `StringPtr(string) *string` - Create string pointer
- `BoolPtr(bool) *bool` - Create bool pointer
- `Int64Ptr(int64) *int64` - Create int64 pointer
- `MustParseUCIInt(interface{}) int64` - Parse int64 (panics on error)

## Usage Example

```go
// Convert from Terraform model to UCI options
var plan networkInterfaceModel
resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

// Convert to UCI options for API call
uciOptions := converters.NetworkInterfaceToOptions(converters.NetworkInterfaceOptions{
    Proto:   &plan.Proto,
    Device:  &plan.Device,
    IPAddr:  &plan.IPAddr,
})

// Call UCI API
secName, err := r.client.UCISection(ctx, "network", "interface", name, uciOptions)

// Convert UCI response back to Terraform model
uciData, err := r.client.UCIGetAll(ctx, "network", name)
model := converters.OptionsToNetworkInterface(uciData)
```

## Testing

Converter tests verify round-trip conversion:
- Convert options struct → UCI map → options struct
- Ensure the final result matches the original

Run tests with:
```bash
go test ./internal/provider/converters/...
```
