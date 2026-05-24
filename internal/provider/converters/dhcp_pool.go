package converters

// DHCPPoolOptions represents DHCP pool configuration options for OpenWrt's UCI DHCP config.
// These options are used when creating or updating DHCP pools via the UCI RPC API.
type DHCPPoolOptions struct {
	// Name is the pool name (typically matches the interface name).
	Name *string
	// Interface specifies which network interface this pool serves.
	Interface *string
	// Start is the first IP address in the DHCP range.
	Start *int64
	// Limit is the number of IP addresses to allocate (total = limit).
	Limit *int64
	// Leasetime specifies DHCP lease duration (e.g., "12h", "24h", "infinite").
	Leasetime *string
	// DHCPv4 sets DHCPv4 mode: "server", "none".
	DHCPv4 *string
	// DHCPv6 sets DHCPv6 mode: "server", "hybrid", "none".
	DHCPv6 *string
	// RA sets Router Advertisement mode: "server", "hybrid", "relay", "disabled".
	RA *string
	// RAFlags specifies RA flags (space-separated).
	RAFlags *string
	// Ignore disables DHCP for this pool when true.
	Ignore *bool
}

// DHCPPoolToOptions converts a DHCPPoolOptions struct to UCI options map.
// Nil pointer fields are omitted from the result.
func DHCPPoolToOptions(m DHCPPoolOptions) map[string]interface{} {
	options := make(map[string]interface{})

	if m.Name != nil {
		options["name"] = *m.Name
	}
	if m.Interface != nil {
		options["interface"] = *m.Interface
	}
	if m.Start != nil {
		options["start"] = *m.Start
	}
	if m.Limit != nil {
		options["limit"] = *m.Limit
	}
	if m.Leasetime != nil {
		options["leasetime"] = *m.Leasetime
	}
	if m.DHCPv4 != nil {
		options["dhcp"] = *m.DHCPv4
	}
	if m.DHCPv6 != nil {
		options["dhcpv6"] = *m.DHCPv6
	}
	if m.RA != nil {
		options["ra"] = *m.RA
	}
	if m.RAFlags != nil {
		options["ra_flags"] = *m.RAFlags
	}
	if m.Ignore != nil && *m.Ignore {
		options["ignore"] = "1"
	}

	return options
}

// OptionsToDHCPPool converts UCI data map to DHCPPoolOptions struct.
// Fields that are missing or invalid in the UCI data are left as nil pointers.
func OptionsToDHCPPool(data map[string]interface{}) DHCPPoolOptions {
	m := DHCPPoolOptions{}

	if v, ok := data["name"].(string); ok {
		m.Name = &v
	}
	if v, ok := data["interface"].(string); ok {
		m.Interface = &v
	}
	if v, ok := data["start"]; ok {
		if val, null := ParseUCIInt(v); !null {
			m.Start = &val
		}
	}
	if v, ok := data["limit"]; ok {
		if val, null := ParseUCIInt(v); !null {
			m.Limit = &val
		}
	}
	if v, ok := data["leasetime"].(string); ok {
		m.Leasetime = &v
	}
	if v, ok := data["dhcp"].(string); ok {
		m.DHCPv4 = &v
	}
	if v, ok := data["dhcpv6"].(string); ok {
		m.DHCPv6 = &v
	}
	if v, ok := data["ra"].(string); ok {
		m.RA = &v
	}
	if v, ok := data["ra_flags"].(string); ok {
		m.RAFlags = &v
	}
	if v, ok := data["ignore"].(string); ok {
		if val, null := ParseUCIBool(v); !null {
			m.Ignore = &val
		}
	}

	return m
}
