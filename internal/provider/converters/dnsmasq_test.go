package converters

import "testing"

func TestDNSMasqToOptions(t *testing.T) {
	tests := []struct {
		name    string
		input   DNSMasqOptions
		wantKey string
		wantVal interface{}
	}{
		{
			name:    "domain needed true",
			input:   DNSMasqOptions{DomainNeeded: ptrBool(true)},
			wantKey: "domainneeded",
			wantVal: "1",
		},
		{
			name:    "domain needed false",
			input:   DNSMasqOptions{DomainNeeded: ptrBool(false)},
			wantKey: "domainneeded",
			wantVal: "0",
		},
		{
			name:    "localise queries",
			input:   DNSMasqOptions{LocaliseQueries: ptrBool(true)},
			wantKey: "localise_queries",
			wantVal: "1",
		},
		{
			name:    "rebind protection",
			input:   DNSMasqOptions{RebindProtection: ptrBool(true)},
			wantKey: "rebind_protection",
			wantVal: "1",
		},
		{
			name:    "local string",
			input:   DNSMasqOptions{Local: ptrString("/lan/")},
			wantKey: "local",
			wantVal: "/lan/",
		},
		{
			name:    "domain string",
			input:   DNSMasqOptions{Domain: ptrString("lan")},
			wantKey: "domain",
			wantVal: "lan",
		},
		{
			name:    "cache size",
			input:   DNSMasqOptions{CacheSize: ptrInt64(1000)},
			wantKey: "cachesize",
			wantVal: int64(1000),
		},
		{
			name:    "authoritative",
			input:   DNSMasqOptions{Authoritative: ptrBool(true)},
			wantKey: "authoritative",
			wantVal: "1",
		},
		{
			name:    "leasefile",
			input:   DNSMasqOptions{LeaseFile: ptrString("/tmp/dhcp.leases")},
			wantKey: "leasefile",
			wantVal: "/tmp/dhcp.leases",
		},
		{
			name:    "resolvfile",
			input:   DNSMasqOptions{ResolvFile: ptrString("/tmp/resolv.conf.d/resolv.conf.auto")},
			wantKey: "resolvfile",
			wantVal: "/tmp/resolv.conf.d/resolv.conf.auto",
		},
		{
			name:    "edns packet max",
			input:   DNSMasqOptions{EDNSPacketMax: ptrInt64(1232)},
			wantKey: "ednspacket_max",
			wantVal: int64(1232),
		},
		{
			name:    "confdir",
			input:   DNSMasqOptions{ConfDir: ptrString("/etc/dnsmasq.d")},
			wantKey: "confdir",
			wantVal: "/etc/dnsmasq.d",
		},
		{
			name:    "rebind domain",
			input:   DNSMasqOptions{RebindDomain: ptrString("whitelist.com")},
			wantKey: "rebind_domain",
			wantVal: "whitelist.com",
		},
		{
			name:    "server",
			input:   DNSMasqOptions{Server: ptrString("8.8.8.8")},
			wantKey: "server",
			wantVal: "8.8.8.8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DNSMasqToOptions(tt.input)
			if val, ok := got[tt.wantKey]; !ok {
				t.Errorf("DNSMasqToOptions() missing key %q", tt.wantKey)
			} else if val != tt.wantVal {
				t.Errorf("DNSMasqToOptions()[%q] = %v, want %v", tt.wantKey, val, tt.wantVal)
			}
		})
	}
}

func TestDNSMasqToOptions_NilFields(t *testing.T) {
	m := DNSMasqOptions{
		DomainNeeded: ptrBool(true),
		Local:        ptrString("/lan/"),
	}
	got := DNSMasqToOptions(m)
	if len(got) != 2 {
		t.Errorf("DNSMasqToOptions() returned %d keys, want 2", len(got))
	}
}

func TestDNSMasqToOptions_AllFields(t *testing.T) {
	m := DNSMasqOptions{
		DomainNeeded:     ptrBool(true),
		LocaliseQueries:  ptrBool(true),
		RebindProtection: ptrBool(true),
		RebindLocalhost:  ptrBool(false),
		Local:            ptrString("/lan/"),
		Domain:           ptrString("lan"),
		ExpandHosts:      ptrBool(true),
		CacheSize:        ptrInt64(1000),
		Authoritative:    ptrBool(true),
		ReadEthers:       ptrBool(false),
		LeaseFile:        ptrString("/tmp/leases"),
		ResolvFile:       ptrString("/tmp/resolv.conf"),
		LocalService:     ptrBool(true),
		EDNSPacketMax:    ptrInt64(1232),
		ConfDir:          ptrString("/etc/dnsmasq.d"),
		RebindDomain:     ptrString("whitelist.com"),
		Server:           ptrString("8.8.8.8"),
	}
	got := DNSMasqToOptions(m)
	if len(got) != 17 {
		t.Errorf("DNSMasqToOptions() returned %d keys, want 17", len(got))
	}
}

func TestOptionsToDNSMasq(t *testing.T) {
	tests := []struct {
		name       string
		data       map[string]interface{}
		checkField string
		wantVal    interface{}
		wantIsNull bool
	}{
		{
			name:       "domain needed 1",
			data:       map[string]interface{}{"domainneeded": "1"},
			checkField: "DomainNeeded",
			wantVal:    true,
			wantIsNull: false,
		},
		{
			name:       "domain needed 0",
			data:       map[string]interface{}{"domainneeded": "0"},
			checkField: "DomainNeeded",
			wantVal:    false,
			wantIsNull: false,
		},
		{
			name:       "domain needed true",
			data:       map[string]interface{}{"domainneeded": "true"},
			checkField: "DomainNeeded",
			wantVal:    true,
			wantIsNull: false,
		},
		{
			name:       "domain needed invalid",
			data:       map[string]interface{}{"domainneeded": "invalid"},
			checkField: "DomainNeeded",
			wantVal:    false,
			wantIsNull: true,
		},
		{
			name:       "localise queries",
			data:       map[string]interface{}{"localise_queries": "1"},
			checkField: "LocaliseQueries",
			wantVal:    true,
			wantIsNull: false,
		},
		{
			name:       "rebind protection",
			data:       map[string]interface{}{"rebind_protection": "1"},
			checkField: "RebindProtection",
			wantVal:    true,
			wantIsNull: false,
		},
		{
			name:       "local",
			data:       map[string]interface{}{"local": "/lan/"},
			checkField: "Local",
			wantVal:    "/lan/",
			wantIsNull: false,
		},
		{
			name:       "domain",
			data:       map[string]interface{}{"domain": "lan"},
			checkField: "Domain",
			wantVal:    "lan",
			wantIsNull: false,
		},
		{
			name:       "expandhosts",
			data:       map[string]interface{}{"expandhosts": "1"},
			checkField: "ExpandHosts",
			wantVal:    true,
			wantIsNull: false,
		},
		{
			name:       "cachesize",
			data:       map[string]interface{}{"cachesize": float64(1000)},
			checkField: "CacheSize",
			wantVal:    int64(1000),
			wantIsNull: false,
		},
		{
			name:       "authoritative",
			data:       map[string]interface{}{"authoritative": "1"},
			checkField: "Authoritative",
			wantVal:    true,
			wantIsNull: false,
		},
		{
			name:       "leasefile",
			data:       map[string]interface{}{"leasefile": "/tmp/leases"},
			checkField: "LeaseFile",
			wantVal:    "/tmp/leases",
			wantIsNull: false,
		},
		{
			name:       "resolvfile",
			data:       map[string]interface{}{"resolvfile": "/tmp/resolv.conf"},
			checkField: "ResolvFile",
			wantVal:    "/tmp/resolv.conf",
			wantIsNull: false,
		},
		{
			name:       "ednspacket_max",
			data:       map[string]interface{}{"ednspacket_max": float64(1232)},
			checkField: "EDNSPacketMax",
			wantVal:    int64(1232),
			wantIsNull: false,
		},
		{
			name:       "confdir",
			data:       map[string]interface{}{"confdir": "/etc/dnsmasq.d"},
			checkField: "ConfDir",
			wantVal:    "/etc/dnsmasq.d",
			wantIsNull: false,
		},
		{
			name:       "rebind_domain",
			data:       map[string]interface{}{"rebind_domain": "whitelist.com"},
			checkField: "RebindDomain",
			wantVal:    "whitelist.com",
			wantIsNull: false,
		},
		{
			name:       "server",
			data:       map[string]interface{}{"server": "8.8.8.8"},
			checkField: "Server",
			wantVal:    "8.8.8.8",
			wantIsNull: false,
		},
		{
			name:       "missing key",
			data:       map[string]interface{}{},
			checkField: "DomainNeeded",
			wantVal:    false,
			wantIsNull: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OptionsToDNSMasq(tt.data)
			checkDNSMasqField(t, got, tt.checkField, tt.wantVal, tt.wantIsNull)
		})
	}
}

func checkDNSMasqField(t *testing.T, m DNSMasqOptions, field string, wantVal interface{}, wantNull bool) {
	t.Helper()
	switch field {
	case "DomainNeeded":
		if wantNull {
			if m.DomainNeeded != nil {
				t.Errorf("OptionsToDNSMasq().DomainNeeded = %v, want nil", *m.DomainNeeded)
			}
		} else {
			if m.DomainNeeded == nil {
				t.Errorf("OptionsToDNSMasq().DomainNeeded = nil, want %v", wantVal)
			} else if *m.DomainNeeded != wantVal {
				t.Errorf("OptionsToDNSMasq().DomainNeeded = %v, want %v", *m.DomainNeeded, wantVal)
			}
		}
	case "LocaliseQueries":
		if wantNull {
			if m.LocaliseQueries != nil {
				t.Errorf("OptionsToDNSMasq().LocaliseQueries = %v, want nil", *m.LocaliseQueries)
			}
		} else {
			if m.LocaliseQueries == nil {
				t.Errorf("OptionsToDNSMasq().LocaliseQueries = nil, want %v", wantVal)
			} else if *m.LocaliseQueries != wantVal {
				t.Errorf("OptionsToDNSMasq().LocaliseQueries = %v, want %v", *m.LocaliseQueries, wantVal)
			}
		}
	case "RebindProtection":
		if wantNull {
			if m.RebindProtection != nil {
				t.Errorf("OptionsToDNSMasq().RebindProtection = %v, want nil", *m.RebindProtection)
			}
		} else {
			if m.RebindProtection == nil {
				t.Errorf("OptionsToDNSMasq().RebindProtection = nil, want %v", wantVal)
			} else if *m.RebindProtection != wantVal {
				t.Errorf("OptionsToDNSMasq().RebindProtection = %v, want %v", *m.RebindProtection, wantVal)
			}
		}
	case "Local":
		if wantNull {
			if m.Local != nil {
				t.Errorf("OptionsToDNSMasq().Local = %v, want nil", *m.Local)
			}
		} else {
			if m.Local == nil {
				t.Errorf("OptionsToDNSMasq().Local = nil, want %v", wantVal)
			} else if *m.Local != wantVal {
				t.Errorf("OptionsToDNSMasq().Local = %v, want %v", *m.Local, wantVal)
			}
		}
	case "Domain":
		if wantNull {
			if m.Domain != nil {
				t.Errorf("OptionsToDNSMasq().Domain = %v, want nil", *m.Domain)
			}
		} else {
			if m.Domain == nil {
				t.Errorf("OptionsToDNSMasq().Domain = nil, want %v", wantVal)
			} else if *m.Domain != wantVal {
				t.Errorf("OptionsToDNSMasq().Domain = %v, want %v", *m.Domain, wantVal)
			}
		}
	case "ExpandHosts":
		if wantNull {
			if m.ExpandHosts != nil {
				t.Errorf("OptionsToDNSMasq().ExpandHosts = %v, want nil", *m.ExpandHosts)
			}
		} else {
			if m.ExpandHosts == nil {
				t.Errorf("OptionsToDNSMasq().ExpandHosts = nil, want %v", wantVal)
			} else if *m.ExpandHosts != wantVal {
				t.Errorf("OptionsToDNSMasq().ExpandHosts = %v, want %v", *m.ExpandHosts, wantVal)
			}
		}
	case "CacheSize":
		if wantNull {
			if m.CacheSize != nil {
				t.Errorf("OptionsToDNSMasq().CacheSize = %v, want nil", *m.CacheSize)
			}
		} else {
			if m.CacheSize == nil {
				t.Errorf("OptionsToDNSMasq().CacheSize = nil, want %v", wantVal)
			} else if *m.CacheSize != wantVal {
				t.Errorf("OptionsToDNSMasq().CacheSize = %v, want %v", *m.CacheSize, wantVal)
			}
		}
	case "Authoritative":
		if wantNull {
			if m.Authoritative != nil {
				t.Errorf("OptionsToDNSMasq().Authoritative = %v, want nil", *m.Authoritative)
			}
		} else {
			if m.Authoritative == nil {
				t.Errorf("OptionsToDNSMasq().Authoritative = nil, want %v", wantVal)
			} else if *m.Authoritative != wantVal {
				t.Errorf("OptionsToDNSMasq().Authoritative = %v, want %v", *m.Authoritative, wantVal)
			}
		}
	case "LeaseFile":
		if wantNull {
			if m.LeaseFile != nil {
				t.Errorf("OptionsToDNSMasq().LeaseFile = %v, want nil", *m.LeaseFile)
			}
		} else {
			if m.LeaseFile == nil {
				t.Errorf("OptionsToDNSMasq().LeaseFile = nil, want %v", wantVal)
			} else if *m.LeaseFile != wantVal {
				t.Errorf("OptionsToDNSMasq().LeaseFile = %v, want %v", *m.LeaseFile, wantVal)
			}
		}
	case "ResolvFile":
		if wantNull {
			if m.ResolvFile != nil {
				t.Errorf("OptionsToDNSMasq().ResolvFile = %v, want nil", *m.ResolvFile)
			}
		} else {
			if m.ResolvFile == nil {
				t.Errorf("OptionsToDNSMasq().ResolvFile = nil, want %v", wantVal)
			} else if *m.ResolvFile != wantVal {
				t.Errorf("OptionsToDNSMasq().ResolvFile = %v, want %v", *m.ResolvFile, wantVal)
			}
		}
	case "EDNSPacketMax":
		if wantNull {
			if m.EDNSPacketMax != nil {
				t.Errorf("OptionsToDNSMasq().EDNSPacketMax = %v, want nil", *m.EDNSPacketMax)
			}
		} else {
			if m.EDNSPacketMax == nil {
				t.Errorf("OptionsToDNSMasq().EDNSPacketMax = nil, want %v", wantVal)
			} else if *m.EDNSPacketMax != wantVal {
				t.Errorf("OptionsToDNSMasq().EDNSPacketMax = %v, want %v", *m.EDNSPacketMax, wantVal)
			}
		}
	case "ConfDir":
		if wantNull {
			if m.ConfDir != nil {
				t.Errorf("OptionsToDNSMasq().ConfDir = %v, want nil", *m.ConfDir)
			}
		} else {
			if m.ConfDir == nil {
				t.Errorf("OptionsToDNSMasq().ConfDir = nil, want %v", wantVal)
			} else if *m.ConfDir != wantVal {
				t.Errorf("OptionsToDNSMasq().ConfDir = %v, want %v", *m.ConfDir, wantVal)
			}
		}
	case "RebindDomain":
		if wantNull {
			if m.RebindDomain != nil {
				t.Errorf("OptionsToDNSMasq().RebindDomain = %v, want nil", *m.RebindDomain)
			}
		} else {
			if m.RebindDomain == nil {
				t.Errorf("OptionsToDNSMasq().RebindDomain = nil, want %v", wantVal)
			} else if *m.RebindDomain != wantVal {
				t.Errorf("OptionsToDNSMasq().RebindDomain = %v, want %v", *m.RebindDomain, wantVal)
			}
		}
	case "Server":
		if wantNull {
			if m.Server != nil {
				t.Errorf("OptionsToDNSMasq().Server = %v, want nil", *m.Server)
			}
		} else {
			if m.Server == nil {
				t.Errorf("OptionsToDNSMasq().Server = nil, want %v", wantVal)
			} else if *m.Server != wantVal {
				t.Errorf("OptionsToDNSMasq().Server = %v, want %v", *m.Server, wantVal)
			}
		}
	}
}

func TestRoundTrip(t *testing.T) {
	original := DNSMasqOptions{
		DomainNeeded:     ptrBool(true),
		LocaliseQueries:  ptrBool(true),
		RebindProtection: ptrBool(false),
		Local:            ptrString("/lan/"),
		Domain:           ptrString("lan"),
		CacheSize:        ptrInt64(1000),
		Authoritative:    ptrBool(true),
		LeaseFile:        ptrString("/tmp/leases"),
		Server:           ptrString("8.8.8.8"),
	}

	options := DNSMasqToOptions(original)
	got := OptionsToDNSMasq(options)

	if original.DomainNeeded == nil != (got.DomainNeeded == nil) {
		t.Error("DomainNeeded mismatch")
	} else if got.DomainNeeded != nil && *got.DomainNeeded != *original.DomainNeeded {
		t.Errorf("DomainNeeded = %v, want %v", *got.DomainNeeded, *original.DomainNeeded)
	}

	if original.Local == nil != (got.Local == nil) {
		t.Error("Local mismatch")
	} else if got.Local != nil && *got.Local != *original.Local {
		t.Errorf("Local = %v, want %v", *got.Local, *original.Local)
	}

	if original.CacheSize == nil != (got.CacheSize == nil) {
		t.Error("CacheSize mismatch")
	} else if got.CacheSize != nil && *got.CacheSize != *original.CacheSize {
		t.Errorf("CacheSize = %v, want %v", *got.CacheSize, *original.CacheSize)
	}
}

func ptrBool(b bool) *bool       { return &b }
func ptrString(s string) *string { return &s }
func ptrInt64(i int64) *int64    { return &i }
