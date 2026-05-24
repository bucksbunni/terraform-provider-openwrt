package converters

// DNSMasqOptions represents DNS/DHCP (dnsmasq) configuration options.
// All fields are pointers to allow distinguishing between unset and empty values.
type DNSMasqOptions struct {
	DomainNeeded     *bool
	LocaliseQueries  *bool
	RebindProtection *bool
	RebindLocalhost  *bool
	Local            *string
	Domain           *string
	ExpandHosts      *bool
	CacheSize        *int64
	Authoritative    *bool
	ReadEthers       *bool
	LeaseFile        *string
	ResolvFile       *string
	LocalService     *bool
	EDNSPacketMax    *int64
	ConfDir          *string
	RebindDomain     *string
	Server           *string
}

// DNSMasqToOptions converts DNSMasqOptions to a UCI options map.
// Pointer fields that are nil are omitted from the result.
func DNSMasqToOptions(m DNSMasqOptions) map[string]interface{} {
	options := make(map[string]interface{})

	if m.DomainNeeded != nil {
		options["domainneeded"] = BoolToUCI(*m.DomainNeeded)
	}
	if m.LocaliseQueries != nil {
		options["localise_queries"] = BoolToUCI(*m.LocaliseQueries)
	}
	if m.RebindProtection != nil {
		options["rebind_protection"] = BoolToUCI(*m.RebindProtection)
	}
	if m.RebindLocalhost != nil {
		options["rebind_localhost"] = BoolToUCI(*m.RebindLocalhost)
	}
	if m.Local != nil {
		options["local"] = *m.Local
	}
	if m.Domain != nil {
		options["domain"] = *m.Domain
	}
	if m.ExpandHosts != nil {
		options["expandhosts"] = BoolToUCI(*m.ExpandHosts)
	}
	if m.CacheSize != nil {
		options["cachesize"] = *m.CacheSize
	}
	if m.Authoritative != nil {
		options["authoritative"] = BoolToUCI(*m.Authoritative)
	}
	if m.ReadEthers != nil {
		options["readethers"] = BoolToUCI(*m.ReadEthers)
	}
	if m.LeaseFile != nil {
		options["leasefile"] = *m.LeaseFile
	}
	if m.ResolvFile != nil {
		options["resolvfile"] = *m.ResolvFile
	}
	if m.LocalService != nil {
		options["localservice"] = BoolToUCI(*m.LocalService)
	}
	if m.EDNSPacketMax != nil {
		options["ednspacket_max"] = *m.EDNSPacketMax
	}
	if m.ConfDir != nil {
		options["confdir"] = *m.ConfDir
	}
	if m.RebindDomain != nil {
		options["rebind_domain"] = *m.RebindDomain
	}
	if m.Server != nil {
		options["server"] = *m.Server
	}

	return options
}

// OptionsToDNSMasq converts a UCI options map to DNSMasqOptions.
func OptionsToDNSMasq(data map[string]interface{}) DNSMasqOptions {
	m := DNSMasqOptions{}

	if v, ok := data["domainneeded"].(string); ok {
		if val, null := ParseUCIBool(v); !null {
			m.DomainNeeded = &val
		}
	}
	if v, ok := data["localise_queries"].(string); ok {
		if val, null := ParseUCIBool(v); !null {
			m.LocaliseQueries = &val
		}
	}
	if v, ok := data["rebind_protection"].(string); ok {
		if val, null := ParseUCIBool(v); !null {
			m.RebindProtection = &val
		}
	}
	if v, ok := data["rebind_localhost"].(string); ok {
		if val, null := ParseUCIBool(v); !null {
			m.RebindLocalhost = &val
		}
	}
	if v, ok := data["local"].(string); ok {
		m.Local = &v
	}
	if v, ok := data["domain"].(string); ok {
		m.Domain = &v
	}
	if v, ok := data["expandhosts"].(string); ok {
		if val, null := ParseUCIBool(v); !null {
			m.ExpandHosts = &val
		}
	}
	if v, ok := data["cachesize"]; ok {
		if val, null := ParseUCIInt(v); !null {
			m.CacheSize = &val
		}
	}
	if v, ok := data["authoritative"].(string); ok {
		if val, null := ParseUCIBool(v); !null {
			m.Authoritative = &val
		}
	}
	if v, ok := data["readethers"].(string); ok {
		if val, null := ParseUCIBool(v); !null {
			m.ReadEthers = &val
		}
	}
	if v, ok := data["leasefile"].(string); ok {
		m.LeaseFile = &v
	}
	if v, ok := data["resolvfile"].(string); ok {
		m.ResolvFile = &v
	}
	if v, ok := data["localservice"].(string); ok {
		if val, null := ParseUCIBool(v); !null {
			m.LocalService = &val
		}
	}
	if v, ok := data["ednspacket_max"]; ok {
		if val, null := ParseUCIInt(v); !null {
			m.EDNSPacketMax = &val
		}
	}
	if v, ok := data["confdir"].(string); ok {
		m.ConfDir = &v
	}
	if v, ok := data["rebind_domain"].(string); ok {
		m.RebindDomain = &v
	}
	if v, ok := data["server"].(string); ok {
		m.Server = &v
	}

	return m
}
