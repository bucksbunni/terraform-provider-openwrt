package converters

// FirewallRuleOptions represents a firewall rule configuration for OpenWrt's UCI firewall config.
// These options are used when creating or updating firewall rules via the UCI RPC API.
type FirewallRuleOptions struct {
	// Name is the unique name of the firewall rule.
	Name *string
	// Src is the source zone name.
	Src *string
	// Dest is the destination zone name.
	Dest *string
	// Proto specifies the protocol: "tcp", "udp", "tcpudp", "icmp", "all".
	Proto *string
	// SrcPort specifies the source port or range (e.g., "1024:65535").
	SrcPort *string
	// DestPort specifies the destination port or range (e.g., "53", "80:443").
	DestPort *string
	// SrcIP specifies source IP address or CIDR.
	SrcIP *string
	// DestIP specifies destination IP address or CIDR.
	DestIP *string
	// Target specifies the action: "ACCEPT", "REJECT", "DROP".
	Target *string
	// Family specifies IP family: "ipv4", "ipv6", "all".
	Family *string
	// ICMPType specifies ICMP types (space-separated).
	ICMPType *string
	// Limit specifies rate limiting (e.g., "1000/sec").
	Limit *string
	// Extra specifies additional iptables options.
	Extra *string
	// Enabled controls whether the rule is active.
	Enabled *bool
}

// FirewallRuleToOptions converts a FirewallRuleOptions struct to UCI options map.
// Nil pointer fields are omitted from the result.
// Note: Enabled=false is represented by omitting the field (UCI default behavior).
func FirewallRuleToOptions(m FirewallRuleOptions) map[string]interface{} {
	options := make(map[string]interface{})

	if m.Name != nil {
		options["name"] = *m.Name
	}
	if m.Src != nil {
		options["src"] = *m.Src
	}
	if m.Dest != nil {
		options["dest"] = *m.Dest
	}
	if m.Proto != nil {
		options["proto"] = *m.Proto
	}
	if m.SrcPort != nil {
		options["src_port"] = *m.SrcPort
	}
	if m.DestPort != nil {
		options["dest_port"] = *m.DestPort
	}
	if m.SrcIP != nil {
		options["src_ip"] = *m.SrcIP
	}
	if m.DestIP != nil {
		options["dest_ip"] = *m.DestIP
	}
	if m.Target != nil {
		options["target"] = *m.Target
	}
	if m.Family != nil {
		options["family"] = *m.Family
	}
	if m.ICMPType != nil {
		options["icmp_type"] = *m.ICMPType
	}
	if m.Limit != nil {
		options["limit"] = *m.Limit
	}
	if m.Extra != nil {
		options["extra"] = *m.Extra
	}
	if m.Enabled != nil && *m.Enabled {
		options["enabled"] = "1"
	}

	return options
}

// OptionsToFirewallRule converts UCI data map to FirewallRuleOptions struct.
// Fields that are missing or invalid in the UCI data are left as nil pointers.
func OptionsToFirewallRule(data map[string]interface{}) FirewallRuleOptions {
	m := FirewallRuleOptions{}

	if v, ok := data["name"].(string); ok {
		m.Name = &v
	}
	if v, ok := data["src"].(string); ok {
		m.Src = &v
	}
	if v, ok := data["dest"].(string); ok {
		m.Dest = &v
	}
	if v, ok := data["proto"].(string); ok {
		m.Proto = &v
	}
	if v, ok := data["src_port"].(string); ok {
		m.SrcPort = &v
	}
	if v, ok := data["dest_port"].(string); ok {
		m.DestPort = &v
	}
	if v, ok := data["src_ip"].(string); ok {
		m.SrcIP = &v
	}
	if v, ok := data["dest_ip"].(string); ok {
		m.DestIP = &v
	}
	if v, ok := data["target"].(string); ok {
		m.Target = &v
	}
	if v, ok := data["family"].(string); ok {
		m.Family = &v
	}
	if v, ok := data["icmp_type"].(string); ok {
		m.ICMPType = &v
	}
	if v, ok := data["limit"].(string); ok {
		m.Limit = &v
	}
	if v, ok := data["extra"].(string); ok {
		m.Extra = &v
	}
	if v, ok := data["enabled"].(string); ok {
		if val, null := ParseUCIBool(v); !null {
			m.Enabled = &val
		}
	}

	return m
}
