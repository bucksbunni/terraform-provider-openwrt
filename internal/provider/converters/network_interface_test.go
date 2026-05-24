package converters

import "testing"

func TestNetworkInterfaceToOptions(t *testing.T) {
	tests := []struct {
		name    string
		input   NetworkInterfaceOptions
		wantKey string
		wantVal interface{}
	}{
		{
			name:    "proto static",
			input:   NetworkInterfaceOptions{Proto: ptrStringP("static")},
			wantKey: "proto",
			wantVal: "static",
		},
		{
			name:    "proto dhcp",
			input:   NetworkInterfaceOptions{Proto: ptrStringP("dhcp")},
			wantKey: "proto",
			wantVal: "dhcp",
		},
		{
			name:    "device",
			input:   NetworkInterfaceOptions{Device: ptrStringP("eth0")},
			wantKey: "device",
			wantVal: "eth0",
		},
		{
			name:    "ipaddr",
			input:   NetworkInterfaceOptions{IPAddr: ptrStringP("192.168.1.1/24")},
			wantKey: "ipaddr",
			wantVal: "192.168.1.1/24",
		},
		{
			name:    "netmask",
			input:   NetworkInterfaceOptions{Netmask: ptrStringP("255.255.255.0")},
			wantKey: "netmask",
			wantVal: "255.255.255.0",
		},
		{
			name:    "gateway",
			input:   NetworkInterfaceOptions{Gateway: ptrStringP("192.168.1.254")},
			wantKey: "gateway",
			wantVal: "192.168.1.254",
		},
		{
			name:    "dns",
			input:   NetworkInterfaceOptions{DNS: ptrStringP("8.8.8.8 8.8.4.4")},
			wantKey: "dns",
			wantVal: "8.8.8.8 8.8.4.4",
		},
		{
			name:    "metric",
			input:   NetworkInterfaceOptions{Metric: ptrInt64P(100)},
			wantKey: "metric",
			wantVal: int64(100),
		},
		{
			name:    "delegate true",
			input:   NetworkInterfaceOptions{Delegate: ptrBoolP(true)},
			wantKey: "delegate",
			wantVal: "1",
		},
		{
			name:    "delegate false",
			input:   NetworkInterfaceOptions{Delegate: ptrBoolP(false)},
			wantKey: "delegate",
			wantVal: "0",
		},
		{
			name:    "ip6addr",
			input:   NetworkInterfaceOptions{IP6Addr: ptrStringP("fd00::1/64")},
			wantKey: "ip6addr",
			wantVal: "fd00::1/64",
		},
		{
			name:    "ip6prefix",
			input:   NetworkInterfaceOptions{IP6Prefix: ptrStringP("fd00::/64")},
			wantKey: "ip6prefix",
			wantVal: "fd00::/64",
		},
		{
			name:    "ip6assign",
			input:   NetworkInterfaceOptions{IP6Assign: ptrStringP("60")},
			wantKey: "ip6assign",
			wantVal: "60",
		},
		{
			name:    "ip6gateway",
			input:   NetworkInterfaceOptions{IP6Gateway: ptrStringP("fd00::ffff")},
			wantKey: "ip6gateway",
			wantVal: "fd00::ffff",
		},
		{
			name:    "auto 1",
			input:   NetworkInterfaceOptions{Auto: ptrStringP("1")},
			wantKey: "auto",
			wantVal: "1",
		},
		{
			name:    "type bridge",
			input:   NetworkInterfaceOptions{IfType: ptrStringP("bridge")},
			wantKey: "type",
			wantVal: "bridge",
		},
		{
			name:    "bridge_empty true",
			input:   NetworkInterfaceOptions{BridgeEmpty: ptrBoolP(true)},
			wantKey: "bridge_empty",
			wantVal: "1",
		},
		{
			name:    "bridge_empty false",
			input:   NetworkInterfaceOptions{BridgeEmpty: ptrBoolP(false)},
			wantKey: "bridge_empty",
			wantVal: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NetworkInterfaceToOptions(tt.input)
			if val, ok := got[tt.wantKey]; !ok {
				t.Errorf("NetworkInterfaceToOptions() missing key %q", tt.wantKey)
			} else if val != tt.wantVal {
				t.Errorf("NetworkInterfaceToOptions()[%q] = %v, want %v", tt.wantKey, val, tt.wantVal)
			}
		})
	}
}

func TestNetworkInterfaceToOptions_NilFields(t *testing.T) {
	m := NetworkInterfaceOptions{
		Proto:  ptrStringP("static"),
		Device: ptrStringP("eth0"),
	}
	got := NetworkInterfaceToOptions(m)
	if len(got) != 2 {
		t.Errorf("NetworkInterfaceToOptions() returned %d keys, want 2", len(got))
	}
}

func TestNetworkInterfaceToOptions_AllFields(t *testing.T) {
	m := NetworkInterfaceOptions{
		Proto:       ptrStringP("static"),
		Device:      ptrStringP("br-lan"),
		IPAddr:      ptrStringP("192.168.1.1/24"),
		Netmask:     ptrStringP("255.255.255.0"),
		Gateway:     ptrStringP("192.168.1.254"),
		DNS:         ptrStringP("8.8.8.8"),
		Metric:      ptrInt64P(100),
		Delegate:    ptrBoolP(true),
		IP6Addr:     ptrStringP("fd00::1/64"),
		IP6Prefix:   ptrStringP("fd00::/64"),
		IP6Assign:   ptrStringP("60"),
		IP6Gateway:  ptrStringP("fd00::ffff"),
		Auto:        ptrStringP("1"),
		IfType:      ptrStringP("bridge"),
		BridgeEmpty: ptrBoolP(false),
	}
	got := NetworkInterfaceToOptions(m)
	if len(got) != 15 {
		t.Errorf("NetworkInterfaceToOptions() returned %d keys, want 15", len(got))
	}
}

func TestOptionsToNetworkInterface(t *testing.T) {
	tests := []struct {
		name       string
		data       map[string]interface{}
		checkField string
		wantVal    interface{}
		wantNull   bool
	}{
		{
			name:       "proto",
			data:       map[string]interface{}{"proto": "static"},
			checkField: "Proto",
			wantVal:    "static",
			wantNull:   false,
		},
		{
			name:       "device",
			data:       map[string]interface{}{"device": "eth0"},
			checkField: "Device",
			wantVal:    "eth0",
			wantNull:   false,
		},
		{
			name:       "ipaddr",
			data:       map[string]interface{}{"ipaddr": "192.168.1.1/24"},
			checkField: "IPAddr",
			wantVal:    "192.168.1.1/24",
			wantNull:   false,
		},
		{
			name:       "netmask",
			data:       map[string]interface{}{"netmask": "255.255.255.0"},
			checkField: "Netmask",
			wantVal:    "255.255.255.0",
			wantNull:   false,
		},
		{
			name:       "gateway",
			data:       map[string]interface{}{"gateway": "192.168.1.254"},
			checkField: "Gateway",
			wantVal:    "192.168.1.254",
			wantNull:   false,
		},
		{
			name:       "dns",
			data:       map[string]interface{}{"dns": "8.8.8.8"},
			checkField: "DNS",
			wantVal:    "8.8.8.8",
			wantNull:   false,
		},
		{
			name:       "metric",
			data:       map[string]interface{}{"metric": float64(100)},
			checkField: "Metric",
			wantVal:    int64(100),
			wantNull:   false,
		},
		{
			name:       "delegate 1",
			data:       map[string]interface{}{"delegate": "1"},
			checkField: "Delegate",
			wantVal:    true,
			wantNull:   false,
		},
		{
			name:       "delegate 0",
			data:       map[string]interface{}{"delegate": "0"},
			checkField: "Delegate",
			wantVal:    false,
			wantNull:   false,
		},
		{
			name:       "ip6addr",
			data:       map[string]interface{}{"ip6addr": "fd00::1/64"},
			checkField: "IP6Addr",
			wantVal:    "fd00::1/64",
			wantNull:   false,
		},
		{
			name:       "auto",
			data:       map[string]interface{}{"auto": "1"},
			checkField: "Auto",
			wantVal:    "1",
			wantNull:   false,
		},
		{
			name:       "type",
			data:       map[string]interface{}{"type": "bridge"},
			checkField: "IfType",
			wantVal:    "bridge",
			wantNull:   false,
		},
		{
			name:       "bridge_empty 1",
			data:       map[string]interface{}{"bridge_empty": "1"},
			checkField: "BridgeEmpty",
			wantVal:    true,
			wantNull:   false,
		},
		{
			name:       "missing key",
			data:       map[string]interface{}{},
			checkField: "Proto",
			wantVal:    "",
			wantNull:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OptionsToNetworkInterface(tt.data)
			checkNetworkInterfaceField(t, got, tt.checkField, tt.wantVal, tt.wantNull)
		})
	}
}

func checkNetworkInterfaceField(t *testing.T, m NetworkInterfaceOptions, field string, wantVal interface{}, wantNull bool) {
	t.Helper()
	switch field {
	case "Proto":
		if wantNull {
			if m.Proto != nil {
				t.Errorf("OptionsToNetworkInterface().Proto = %v, want nil", *m.Proto)
			}
		} else {
			if m.Proto == nil || *m.Proto != wantVal {
				t.Errorf("OptionsToNetworkInterface().Proto = %v, want %v", m.Proto, wantVal)
			}
		}
	case "Device":
		if wantNull {
			if m.Device != nil {
				t.Errorf("OptionsToNetworkInterface().Device = %v, want nil", *m.Device)
			}
		} else {
			if m.Device == nil || *m.Device != wantVal {
				t.Errorf("OptionsToNetworkInterface().Device = %v, want %v", m.Device, wantVal)
			}
		}
	case "IPAddr":
		if wantNull {
			if m.IPAddr != nil {
				t.Errorf("OptionsToNetworkInterface().IPAddr = %v, want nil", *m.IPAddr)
			}
		} else {
			if m.IPAddr == nil || *m.IPAddr != wantVal {
				t.Errorf("OptionsToNetworkInterface().IPAddr = %v, want %v", m.IPAddr, wantVal)
			}
		}
	case "Gateway":
		if wantNull {
			if m.Gateway != nil {
				t.Errorf("OptionsToNetworkInterface().Gateway = %v, want nil", *m.Gateway)
			}
		} else {
			if m.Gateway == nil || *m.Gateway != wantVal {
				t.Errorf("OptionsToNetworkInterface().Gateway = %v, want %v", m.Gateway, wantVal)
			}
		}
	case "DNS":
		if wantNull {
			if m.DNS != nil {
				t.Errorf("OptionsToNetworkInterface().DNS = %v, want nil", *m.DNS)
			}
		} else {
			if m.DNS == nil || *m.DNS != wantVal {
				t.Errorf("OptionsToNetworkInterface().DNS = %v, want %v", m.DNS, wantVal)
			}
		}
	case "Metric":
		if wantNull {
			if m.Metric != nil {
				t.Errorf("OptionsToNetworkInterface().Metric = %v, want nil", *m.Metric)
			}
		} else {
			if m.Metric == nil || *m.Metric != wantVal {
				t.Errorf("OptionsToNetworkInterface().Metric = %v, want %v", m.Metric, wantVal)
			}
		}
	case "Delegate":
		if wantNull {
			if m.Delegate != nil {
				t.Errorf("OptionsToNetworkInterface().Delegate = %v, want nil", *m.Delegate)
			}
		} else {
			if m.Delegate == nil || *m.Delegate != wantVal {
				t.Errorf("OptionsToNetworkInterface().Delegate = %v, want %v", m.Delegate, wantVal)
			}
		}
	case "IP6Addr":
		if wantNull {
			if m.IP6Addr != nil {
				t.Errorf("OptionsToNetworkInterface().IP6Addr = %v, want nil", *m.IP6Addr)
			}
		} else {
			if m.IP6Addr == nil || *m.IP6Addr != wantVal {
				t.Errorf("OptionsToNetworkInterface().IP6Addr = %v, want %v", m.IP6Addr, wantVal)
			}
		}
	case "Auto":
		if wantNull {
			if m.Auto != nil {
				t.Errorf("OptionsToNetworkInterface().Auto = %v, want nil", *m.Auto)
			}
		} else {
			if m.Auto == nil || *m.Auto != wantVal {
				t.Errorf("OptionsToNetworkInterface().Auto = %v, want %v", m.Auto, wantVal)
			}
		}
	case "IfType":
		if wantNull {
			if m.IfType != nil {
				t.Errorf("OptionsToNetworkInterface().IfType = %v, want nil", *m.IfType)
			}
		} else {
			if m.IfType == nil || *m.IfType != wantVal {
				t.Errorf("OptionsToNetworkInterface().IfType = %v, want %v", m.IfType, wantVal)
			}
		}
	case "BridgeEmpty":
		if wantNull {
			if m.BridgeEmpty != nil {
				t.Errorf("OptionsToNetworkInterface().BridgeEmpty = %v, want nil", *m.BridgeEmpty)
			}
		} else {
			if m.BridgeEmpty == nil || *m.BridgeEmpty != wantVal {
				t.Errorf("OptionsToNetworkInterface().BridgeEmpty = %v, want %v", m.BridgeEmpty, wantVal)
			}
		}
	}
}

func TestNetworkInterfaceRoundTrip(t *testing.T) {
	original := NetworkInterfaceOptions{
		Proto:    ptrStringP("static"),
		Device:   ptrStringP("eth0"),
		IPAddr:   ptrStringP("192.168.1.1/24"),
		Gateway:  ptrStringP("192.168.1.254"),
		Metric:   ptrInt64P(100),
		Delegate: ptrBoolP(true),
	}

	options := NetworkInterfaceToOptions(original)
	got := OptionsToNetworkInterface(options)

	if original.Proto == nil != (got.Proto == nil) {
		t.Error("Proto mismatch")
	} else if got.Proto != nil && *got.Proto != *original.Proto {
		t.Errorf("Proto = %v, want %v", *got.Proto, *original.Proto)
	}

	if original.Device == nil != (got.Device == nil) {
		t.Error("Device mismatch")
	} else if got.Device != nil && *got.Device != *original.Device {
		t.Errorf("Device = %v, want %v", *got.Device, *original.Device)
	}

	if original.Metric == nil != (got.Metric == nil) {
		t.Error("Metric mismatch")
	} else if got.Metric != nil && *got.Metric != *original.Metric {
		t.Errorf("Metric = %v, want %v", *got.Metric, *original.Metric)
	}
}

func ptrStringP(s string) *string { return &s }
func ptrBoolP(b bool) *bool       { return &b }
func ptrInt64P(i int64) *int64    { return &i }
