package converters

import "testing"

func TestFirewallZoneToOptions(t *testing.T) {
	tests := []struct {
		name    string
		input   FirewallZoneOptions
		wantKey string
		wantVal interface{}
	}{
		{
			name:    "name",
			input:   FirewallZoneOptions{Name: ptrFWS("wan")},
			wantKey: "name",
			wantVal: "wan",
		},
		{
			name:    "input ACCEPT",
			input:   FirewallZoneOptions{Input: ptrFWS("ACCEPT")},
			wantKey: "input",
			wantVal: "ACCEPT",
		},
		{
			name:    "output REJECT",
			input:   FirewallZoneOptions{Output: ptrFWS("REJECT")},
			wantKey: "output",
			wantVal: "REJECT",
		},
		{
			name:    "forward DROP",
			input:   FirewallZoneOptions{Forward: ptrFWS("DROP")},
			wantKey: "forward",
			wantVal: "DROP",
		},
		{
			name:    "masq true",
			input:   FirewallZoneOptions{Masq: ptrFWB(true)},
			wantKey: "masq",
			wantVal: "1",
		},
		{
			name:    "masq false (not set)",
			input:   FirewallZoneOptions{Masq: ptrFWB(false)},
			wantKey: "masq",
			wantVal: nil,
		},
		{
			name:    "masq_src",
			input:   FirewallZoneOptions{MasqSrc: ptrFWS("192.168.0.0/16")},
			wantKey: "masq_src",
			wantVal: "192.168.0.0/16",
		},
		{
			name:    "masq_dest",
			input:   FirewallZoneOptions{MasqDest: ptrFWS("10.0.0.0/8")},
			wantKey: "masq_dest",
			wantVal: "10.0.0.0/8",
		},
		{
			name:    "mtu_fix true",
			input:   FirewallZoneOptions{MtuFix: ptrFWB(true)},
			wantKey: "mtu_fix",
			wantVal: "1",
		},
		{
			name:    "mtu_fix false (not set)",
			input:   FirewallZoneOptions{MtuFix: ptrFWB(false)},
			wantKey: "mtu_fix",
			wantVal: nil,
		},
		{
			name:    "network",
			input:   FirewallZoneOptions{Network: ptrFWS("wan")},
			wantKey: "network",
			wantVal: "wan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FirewallZoneToOptions(tt.input)
			if tt.wantVal == nil {
				if _, ok := got[tt.wantKey]; ok {
					t.Errorf("FirewallZoneToOptions()[%q] should not be set", tt.wantKey)
				}
			} else if val, ok := got[tt.wantKey]; !ok {
				t.Errorf("FirewallZoneToOptions() missing key %q", tt.wantKey)
			} else if val != tt.wantVal {
				t.Errorf("FirewallZoneToOptions()[%q] = %v, want %v", tt.wantKey, val, tt.wantVal)
			}
		})
	}
}

func TestFirewallZoneToOptions_AllFields(t *testing.T) {
	m := FirewallZoneOptions{
		Name:     ptrFWS("wan"),
		Input:    ptrFWS("ACCEPT"),
		Output:   ptrFWS("ACCEPT"),
		Forward:  ptrFWS("REJECT"),
		Masq:     ptrFWB(true),
		MasqSrc:  ptrFWS("192.168.0.0/16"),
		MasqDest: ptrFWS("10.0.0.0/8"),
		MtuFix:   ptrFWB(true),
		Network:  ptrFWS("wan"),
	}
	got := FirewallZoneToOptions(m)
	if len(got) != 9 {
		t.Errorf("FirewallZoneToOptions() returned %d keys, want 9", len(got))
	}
}

func TestOptionsToFirewallZone(t *testing.T) {
	tests := []struct {
		name       string
		data       map[string]interface{}
		checkField string
		wantVal    interface{}
		wantNull   bool
	}{
		{
			name:       "name",
			data:       map[string]interface{}{"name": "wan"},
			checkField: "Name",
			wantVal:    "wan",
			wantNull:   false,
		},
		{
			name:       "input",
			data:       map[string]interface{}{"input": "ACCEPT"},
			checkField: "Input",
			wantVal:    "ACCEPT",
			wantNull:   false,
		},
		{
			name:       "output",
			data:       map[string]interface{}{"output": "REJECT"},
			checkField: "Output",
			wantVal:    "REJECT",
			wantNull:   false,
		},
		{
			name:       "forward",
			data:       map[string]interface{}{"forward": "DROP"},
			checkField: "Forward",
			wantVal:    "DROP",
			wantNull:   false,
		},
		{
			name:       "masq 1",
			data:       map[string]interface{}{"masq": "1"},
			checkField: "Masq",
			wantVal:    true,
			wantNull:   false,
		},
		{
			name:       "masq 0",
			data:       map[string]interface{}{"masq": "0"},
			checkField: "Masq",
			wantVal:    false,
			wantNull:   false,
		},
		{
			name:       "masq_src",
			data:       map[string]interface{}{"masq_src": "192.168.0.0/16"},
			checkField: "MasqSrc",
			wantVal:    "192.168.0.0/16",
			wantNull:   false,
		},
		{
			name:       "masq_dest",
			data:       map[string]interface{}{"masq_dest": "10.0.0.0/8"},
			checkField: "MasqDest",
			wantVal:    "10.0.0.0/8",
			wantNull:   false,
		},
		{
			name:       "mtu_fix 1",
			data:       map[string]interface{}{"mtu_fix": "1"},
			checkField: "MtuFix",
			wantVal:    true,
			wantNull:   false,
		},
		{
			name:       "network",
			data:       map[string]interface{}{"network": "wan"},
			checkField: "Network",
			wantVal:    "wan",
			wantNull:   false,
		},
		{
			name:       "missing name",
			data:       map[string]interface{}{},
			checkField: "Name",
			wantVal:    "",
			wantNull:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OptionsToFirewallZone(tt.data)
			checkFirewallZoneField(t, got, tt.checkField, tt.wantVal, tt.wantNull)
		})
	}
}

func checkFirewallZoneField(t *testing.T, m FirewallZoneOptions, field string, wantVal interface{}, wantNull bool) {
	t.Helper()
	switch field {
	case "Name":
		if wantNull {
			if m.Name != nil {
				t.Errorf("OptionsToFirewallZone().Name = %v, want nil", *m.Name)
			}
		} else {
			if m.Name == nil || *m.Name != wantVal {
				t.Errorf("OptionsToFirewallZone().Name = %v, want %v", m.Name, wantVal)
			}
		}
	case "Input":
		if wantNull {
			if m.Input != nil {
				t.Errorf("OptionsToFirewallZone().Input = %v, want nil", *m.Input)
			}
		} else {
			if m.Input == nil || *m.Input != wantVal {
				t.Errorf("OptionsToFirewallZone().Input = %v, want %v", m.Input, wantVal)
			}
		}
	case "Output":
		if wantNull {
			if m.Output != nil {
				t.Errorf("OptionsToFirewallZone().Output = %v, want nil", *m.Output)
			}
		} else {
			if m.Output == nil || *m.Output != wantVal {
				t.Errorf("OptionsToFirewallZone().Output = %v, want %v", m.Output, wantVal)
			}
		}
	case "Forward":
		if wantNull {
			if m.Forward != nil {
				t.Errorf("OptionsToFirewallZone().Forward = %v, want nil", *m.Forward)
			}
		} else {
			if m.Forward == nil || *m.Forward != wantVal {
				t.Errorf("OptionsToFirewallZone().Forward = %v, want %v", m.Forward, wantVal)
			}
		}
	case "Masq":
		if wantNull {
			if m.Masq != nil {
				t.Errorf("OptionsToFirewallZone().Masq = %v, want nil", *m.Masq)
			}
		} else {
			if m.Masq == nil || *m.Masq != wantVal {
				t.Errorf("OptionsToFirewallZone().Masq = %v, want %v", m.Masq, wantVal)
			}
		}
	case "MasqSrc":
		if wantNull {
			if m.MasqSrc != nil {
				t.Errorf("OptionsToFirewallZone().MasqSrc = %v, want nil", *m.MasqSrc)
			}
		} else {
			if m.MasqSrc == nil || *m.MasqSrc != wantVal {
				t.Errorf("OptionsToFirewallZone().MasqSrc = %v, want %v", m.MasqSrc, wantVal)
			}
		}
	case "MasqDest":
		if wantNull {
			if m.MasqDest != nil {
				t.Errorf("OptionsToFirewallZone().MasqDest = %v, want nil", *m.MasqDest)
			}
		} else {
			if m.MasqDest == nil || *m.MasqDest != wantVal {
				t.Errorf("OptionsToFirewallZone().MasqDest = %v, want %v", m.MasqDest, wantVal)
			}
		}
	case "MtuFix":
		if wantNull {
			if m.MtuFix != nil {
				t.Errorf("OptionsToFirewallZone().MtuFix = %v, want nil", *m.MtuFix)
			}
		} else {
			if m.MtuFix == nil || *m.MtuFix != wantVal {
				t.Errorf("OptionsToFirewallZone().MtuFix = %v, want %v", m.MtuFix, wantVal)
			}
		}
	case "Network":
		if wantNull {
			if m.Network != nil {
				t.Errorf("OptionsToFirewallZone().Network = %v, want nil", *m.Network)
			}
		} else {
			if m.Network == nil || *m.Network != wantVal {
				t.Errorf("OptionsToFirewallZone().Network = %v, want %v", m.Network, wantVal)
			}
		}
	}
}

func TestFirewallZoneRoundTrip(t *testing.T) {
	original := FirewallZoneOptions{
		Name:    ptrFWS("wan"),
		Input:   ptrFWS("ACCEPT"),
		Masq:    ptrFWB(true),
		MasqSrc: ptrFWS("192.168.0.0/16"),
		Network: ptrFWS("wan"),
	}

	options := FirewallZoneToOptions(original)
	got := OptionsToFirewallZone(options)

	if original.Name == nil != (got.Name == nil) {
		t.Error("Name mismatch")
	} else if got.Name != nil && *got.Name != *original.Name {
		t.Errorf("Name = %v, want %v", *got.Name, *original.Name)
	}

	if original.Masq == nil != (got.Masq == nil) {
		t.Error("Masq mismatch")
	} else if got.Masq != nil && *got.Masq != *original.Masq {
		t.Errorf("Masq = %v, want %v", *got.Masq, *original.Masq)
	}
}

func ptrFWS(s string) *string { return &s }
func ptrFWB(b bool) *bool     { return &b }
