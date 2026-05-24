package converters

// DHCPHostOptions represents a static DHCP host reservation for OpenWrt's UCI DHCP config.
// These options are used when creating or updating DHCP hosts via the UCI RPC API.
type DHCPHostOptions struct {
	// Name is the hostname for this reservation.
	Name *string
	// IP is the reserved IP address.
	IP *string
	// MAC is the MAC address to match.
	MAC *string
	// Leasetime overrides the default lease time for this host.
	Leasetime *string
}

// DHCPHostToOptions converts a DHCPHostOptions struct to UCI options map.
// Nil pointer fields are omitted from the result.
func DHCPHostToOptions(m DHCPHostOptions) map[string]interface{} {
	options := make(map[string]interface{})

	if m.Name != nil {
		options["name"] = *m.Name
	}
	if m.IP != nil {
		options["ip"] = *m.IP
	}
	if m.MAC != nil {
		options["mac"] = *m.MAC
	}
	if m.Leasetime != nil {
		options["leasetime"] = *m.Leasetime
	}

	return options
}

// OptionsToDHCPHost converts UCI data map to DHCPHostOptions struct.
// Fields that are missing or invalid in the UCI data are left as nil pointers.
func OptionsToDHCPHost(data map[string]interface{}) DHCPHostOptions {
	m := DHCPHostOptions{}

	if v, ok := data["name"].(string); ok {
		m.Name = &v
	}
	if v, ok := data["ip"].(string); ok {
		m.IP = &v
	}
	if v, ok := data["mac"].(string); ok {
		m.MAC = &v
	}
	if v, ok := data["leasetime"].(string); ok {
		m.Leasetime = &v
	}

	return m
}
