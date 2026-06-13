package provider

import "testing"

func TestExtractInterfaceName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"wgclient1", "wgclient1"},
		{"wg0", "wg0"},
		{"wg_vpn", "vpn"},
		{"simple", "simple"},
		{"wgtest", "wgtest"},
		{"wgvpn_client", "client"},
	}

	for _, tt := range tests {
		result := extractInterfaceName(tt.input)
		if result != tt.expected {
			t.Errorf("extractInterfaceName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
