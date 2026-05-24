package importid

import "testing"

func TestParseUCISectionID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantCfg string
		wantNm  string
		wantErr bool
	}{
		{"valid network/lan", "network/lan", "network", "lan", false},
		{"valid dhcp/pool1", "dhcp/pool1", "dhcp", "pool1", false},
		{"valid firewall/wan", "firewall/wan", "firewall", "wan", false},
		{"valid with underscore", "network/wg0_client1", "network", "wg0_client1", false},
		{"valid with numbers", "network/eth0.100", "network", "eth0.100", false},
		{"empty string", "", "", "", true},
		{"no slash", "invalid", "", "", true},
		{"only slash", "/", "", "", true},
		{"slash at start", "/lan", "", "", true},
		{"slash at end", "network/", "", "", true},
		{"double slash", "network//lan", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, nm, err := ParseUCISectionID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUCISectionID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if cfg != tt.wantCfg {
				t.Errorf("ParseUCISectionID() config = %v, want %v", cfg, tt.wantCfg)
			}
			if nm != tt.wantNm {
				t.Errorf("ParseUCISectionID() name = %v, want %v", nm, tt.wantNm)
			}
		})
	}
}

func TestFormatUCISectionID(t *testing.T) {
	tests := []struct {
		config string
		name   string
		want   string
	}{
		{"network", "lan", "network/lan"},
		{"dhcp", "pool1", "dhcp/pool1"},
		{"firewall", "wan", "firewall/wan"},
		{"wireless", "wwan0", "wireless/wwan0"},
	}

	for _, tt := range tests {
		t.Run(tt.config+"/"+tt.name, func(t *testing.T) {
			if got := FormatUCISectionID(tt.config, tt.name); got != tt.want {
				t.Errorf("FormatUCISectionID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseNetworkInterfaceID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		want    string
		wantErr bool
	}{
		{"valid", "network/lan", "lan", false},
		{"valid with eth", "network/eth0", "eth0", false},
		{"valid with wg", "network/wg0", "wg0", false},
		{"wrong config", "dhcp/lan", "", true},
		{"empty", "", "", true},
		{"no slash", "lan", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseNetworkInterfaceID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNetworkInterfaceID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseNetworkInterfaceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatNetworkInterfaceID(t *testing.T) {
	if got := FormatNetworkInterfaceID("lan"); got != "network/lan" {
		t.Errorf("FormatNetworkInterfaceID() = %v, want network/lan", got)
	}
}

func TestParseDHCPPoolID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		want    string
		wantErr bool
	}{
		{"valid", "dhcp/pool1", "pool1", false},
		{"wrong config", "network/pool1", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDHCPPoolID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDHCPPoolID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDHCPPoolID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatDHCPPoolID(t *testing.T) {
	if got := FormatDHCPPoolID("pool1"); got != "dhcp/pool1" {
		t.Errorf("FormatDHCPPoolID() = %v, want dhcp/pool1", got)
	}
}

func TestParseFirewallZoneID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		want    string
		wantErr bool
	}{
		{"valid", "firewall/wan", "wan", false},
		{"wrong config", "network/wan", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFirewallZoneID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFirewallZoneID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFirewallZoneID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatFirewallZoneID(t *testing.T) {
	if got := FormatFirewallZoneID("wan"); got != "firewall/wan" {
		t.Errorf("FormatFirewallZoneID() = %v, want firewall/wan", got)
	}
}

func TestParseFirewallRuleID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		want    string
		wantErr bool
	}{
		{"valid", "firewall/allow_wan", "allow_wan", false},
		{"wrong config", "network/allow_wan", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFirewallRuleID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFirewallRuleID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFirewallRuleID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatFirewallRuleID(t *testing.T) {
	if got := FormatFirewallRuleID("allow_wan"); got != "firewall/allow_wan" {
		t.Errorf("FormatFirewallRuleID() = %v, want firewall/allow_wan", got)
	}
}

func TestParseFirewallForwardingID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantSrc string
		wantDst string
		wantErr bool
	}{
		{"valid wan_to_lan", "firewall/wan_lan", "wan", "lan", false},
		{"valid lan_to_wan", "firewall/lan_wan", "lan", "wan", false},
		{"valid with zone names", "firewall/office_vpn", "office", "vpn", false},
		{"no underscore", "firewall/wan", "", "", true},
		{"only underscore", "firewall/_", "", "", true},
		{"empty", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, dst, err := ParseFirewallForwardingID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFirewallForwardingID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if src != tt.wantSrc {
				t.Errorf("ParseFirewallForwardingID() src = %v, want %v", src, tt.wantSrc)
			}
			if dst != tt.wantDst {
				t.Errorf("ParseFirewallForwardingID() dst = %v, want %v", dst, tt.wantDst)
			}
		})
	}
}

func TestFormatFirewallForwardingID(t *testing.T) {
	if got := FormatFirewallForwardingID("wan", "lan"); got != "firewall/wan_lan" {
		t.Errorf("FormatFirewallForwardingID() = %v, want firewall/wan_lan", got)
	}
}

func TestParseWirelessInterfaceID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		want    string
		wantErr bool
	}{
		{"valid", "wireless/wwan0", "wwan0", false},
		{"valid with radio", "wireless/radio0.network1", "radio0.network1", false},
		{"wrong config", "network/wwan0", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseWirelessInterfaceID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseWirelessInterfaceID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseWirelessInterfaceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatWirelessInterfaceID(t *testing.T) {
	if got := FormatWirelessInterfaceID("wwan0"); got != "wireless/wwan0" {
		t.Errorf("FormatWirelessInterfaceID() = %v, want wireless/wwan0", got)
	}
}

func TestParseDropbearID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		want    string
		wantErr bool
	}{
		{"valid", "dropbear/main", "main", false},
		{"wrong config", "network/main", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDropbearID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDropbearID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDropbearID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatDropbearID(t *testing.T) {
	if got := FormatDropbearID("main"); got != "dropbear/main" {
		t.Errorf("FormatDropbearID() = %v, want dropbear/main", got)
	}
}

func TestParseSystemLEDID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		want    string
		wantErr bool
	}{
		{"valid", "system/led_wan", "led_wan", false},
		{"wrong config", "network/led_wan", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSystemLEDID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSystemLEDID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseSystemLEDID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatSystemLEDID(t *testing.T) {
	if got := FormatSystemLEDID("led_wan"); got != "system/led_wan" {
		t.Errorf("FormatSystemLEDID() = %v, want system/led_wan", got)
	}
}
