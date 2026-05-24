package converters

// FirewallZoneOptions represents firewall zone configuration options.
// All fields are pointers to allow distinguishing between unset and empty values.
type FirewallZoneOptions struct {
	Name     *string
	Input    *string
	Output   *string
	Forward  *string
	Masq     *bool
	MasqSrc  *string
	MasqDest *string
	MtuFix   *bool
	Network  *string
}

// FirewallZoneToOptions converts FirewallZoneOptions to a UCI options map.
// Pointer fields that are nil are omitted from the result.
func FirewallZoneToOptions(m FirewallZoneOptions) map[string]interface{} {
	options := make(map[string]interface{})

	if m.Name != nil {
		options["name"] = *m.Name
	}
	if m.Input != nil {
		options["input"] = *m.Input
	}
	if m.Output != nil {
		options["output"] = *m.Output
	}
	if m.Forward != nil {
		options["forward"] = *m.Forward
	}
	if m.Masq != nil {
		if *m.Masq {
			options["masq"] = "1"
		}
	}
	if m.MasqSrc != nil {
		options["masq_src"] = *m.MasqSrc
	}
	if m.MasqDest != nil {
		options["masq_dest"] = *m.MasqDest
	}
	if m.MtuFix != nil {
		if *m.MtuFix {
			options["mtu_fix"] = "1"
		}
	}
	if m.Network != nil {
		options["network"] = *m.Network
	}

	return options
}

// OptionsToFirewallZone converts a UCI options map to FirewallZoneOptions.
func OptionsToFirewallZone(data map[string]interface{}) FirewallZoneOptions {
	m := FirewallZoneOptions{}

	if v, ok := data["name"].(string); ok {
		m.Name = &v
	}
	if v, ok := data["input"].(string); ok {
		m.Input = &v
	}
	if v, ok := data["output"].(string); ok {
		m.Output = &v
	}
	if v, ok := data["forward"].(string); ok {
		m.Forward = &v
	}
	if v, ok := data["masq"].(string); ok {
		if val, null := ParseUCIBool(v); !null {
			m.Masq = &val
		}
	}
	if v, ok := data["masq_src"].(string); ok {
		m.MasqSrc = &v
	}
	if v, ok := data["masq_dest"].(string); ok {
		m.MasqDest = &v
	}
	if v, ok := data["mtu_fix"].(string); ok {
		if val, null := ParseUCIBool(v); !null {
			m.MtuFix = &val
		}
	}
	if v, ok := data["network"].(string); ok {
		m.Network = &v
	}

	return m
}
