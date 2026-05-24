package converters

// FirewallForwardingOptions represents a firewall forwarding rule for OpenWrt's UCI firewall config.
// Forwarding rules allow traffic between zones. The ID format is "src_dest".
type FirewallForwardingOptions struct {
	// Src is the source zone name.
	Src *string
	// Dest is the destination zone name.
	Dest *string
}

// FirewallForwardingToOptions converts a FirewallForwardingOptions struct to UCI options map.
// Nil pointer fields are omitted from the result.
func FirewallForwardingToOptions(m FirewallForwardingOptions) map[string]interface{} {
	options := make(map[string]interface{})

	if m.Src != nil {
		options["src"] = *m.Src
	}
	if m.Dest != nil {
		options["dest"] = *m.Dest
	}

	return options
}

// OptionsToFirewallForwarding converts UCI data map to FirewallForwardingOptions struct.
// Fields that are missing or invalid in the UCI data are left as nil pointers.
func OptionsToFirewallForwarding(data map[string]interface{}) FirewallForwardingOptions {
	m := FirewallForwardingOptions{}

	if v, ok := data["src"].(string); ok {
		m.Src = &v
	}
	if v, ok := data["dest"].(string); ok {
		m.Dest = &v
	}

	return m
}
