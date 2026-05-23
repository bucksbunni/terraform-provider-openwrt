package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestNetworkInterfaceResource_Schema(t *testing.T) {
	r := NewNetworkInterfaceResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages an OpenWrt network interface." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "name", "proto", "device", "ipaddr", "netmask", "gateway", "dns", "metric", "delegate", "ip6addr", "ip6prefix", "ip6assign", "ip6gateway", "auto", "type", "bridge_empty"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestNetworkInterfaceResource_Metadata(t *testing.T) {
	r := NewNetworkInterfaceResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_network_interface" {
		t.Fatalf("expected type name 'openwrt_network_interface', got %q", resp.TypeName)
	}
}

func TestNetworkInterfaceResource_ImplementsResource(t *testing.T) {
	r := NewNetworkInterfaceResource()
	var _ resource.Resource = r
}

func TestNetworkDeviceResource_Schema(t *testing.T) {
	r := NewNetworkDeviceResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages an OpenWrt network device (bridge, bonding, etc.)." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "name", "type", "ports", "policy", "xmit_hash_policy"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestNetworkDeviceResource_Metadata(t *testing.T) {
	r := NewNetworkDeviceResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_network_device" {
		t.Fatalf("expected type name 'openwrt_network_device', got %q", resp.TypeName)
	}
}

func TestNetworkDeviceResource_ImplementsResource(t *testing.T) {
	r := NewNetworkDeviceResource()
	var _ resource.Resource = r
}

func TestNetworkBridgeVlanResource_Schema(t *testing.T) {
	r := NewNetworkBridgeVlanResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages a bridge VLAN assignment in OpenWrt." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "device", "vlan", "ports"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestNetworkBridgeVlanResource_Metadata(t *testing.T) {
	r := NewNetworkBridgeVlanResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_network_bridge_vlan" {
		t.Fatalf("expected type name 'openwrt_network_bridge_vlan', got %q", resp.TypeName)
	}
}

func TestNetworkBridgeVlanResource_ImplementsResource(t *testing.T) {
	r := NewNetworkBridgeVlanResource()
	var _ resource.Resource = r
}

func TestNetworkGlobalsResource_Schema(t *testing.T) {
	r := NewNetworkGlobalsResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages OpenWrt network global settings." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "ula_prefix", "packet_steering"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestNetworkGlobalsResource_Metadata(t *testing.T) {
	r := NewNetworkGlobalsResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_network_globals" {
		t.Fatalf("expected type name 'openwrt_network_globals', got %q", resp.TypeName)
	}
}

func TestNetworkGlobalsResource_ImplementsResource(t *testing.T) {
	r := NewNetworkGlobalsResource()
	var _ resource.Resource = r
}

func TestNetworkWireguardResource_Schema(t *testing.T) {
	r := NewNetworkWireguardResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages a WireGuard peer in OpenWrt." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "name", "description", "public_key", "endpoint_host", "endpoint_port", "persistent_keepalive", "allowed_ips"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestNetworkWireguardResource_Metadata(t *testing.T) {
	r := NewNetworkWireguardResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_network_wireguard" {
		t.Fatalf("expected type name 'openwrt_network_wireguard', got %q", resp.TypeName)
	}
}

func TestNetworkWireguardResource_ImplementsResource(t *testing.T) {
	r := NewNetworkWireguardResource()
	var _ resource.Resource = r
}

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
