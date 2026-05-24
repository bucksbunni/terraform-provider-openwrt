package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestFirewallZoneResource_Schema(t *testing.T) {
	r := NewFirewallZoneResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages an OpenWrt firewall zone." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "name", "input", "output", "forward", "masq", "masq_src", "masq_dest", "mtu_fix", "network"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestFirewallZoneResource_Metadata(t *testing.T) {
	r := NewFirewallZoneResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_firewall_zone" {
		t.Fatalf("expected type name 'openwrt_firewall_zone', got %q", resp.TypeName)
	}
}

func TestFirewallZoneResource_ImplementsResource(t *testing.T) {
	r := NewFirewallZoneResource()
	var _ resource.Resource = r
}

func TestFirewallRuleResource_Schema(t *testing.T) {
	r := NewFirewallRuleResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages an OpenWrt firewall rule." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "name", "src", "dest", "proto", "src_port", "dest_port", "src_ip", "dest_ip", "target", "family", "icmp_type", "limit", "extra", "enabled"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestFirewallRuleResource_Metadata(t *testing.T) {
	r := NewFirewallRuleResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_firewall_rule" {
		t.Fatalf("expected type name 'openwrt_firewall_rule', got %q", resp.TypeName)
	}
}

func TestFirewallRuleResource_ImplementsResource(t *testing.T) {
	r := NewFirewallRuleResource()
	var _ resource.Resource = r
}

func TestFirewallForwardingResource_Schema(t *testing.T) {
	r := NewFirewallForwardingResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages an OpenWrt firewall forwarding rule (zone-to-zone traffic)." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "src", "dest"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestFirewallForwardingResource_Metadata(t *testing.T) {
	r := NewFirewallForwardingResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_firewall_forwarding" {
		t.Fatalf("expected type name 'openwrt_firewall_forwarding', got %q", resp.TypeName)
	}
}

func TestFirewallForwardingResource_ImplementsResource(t *testing.T) {
	r := NewFirewallForwardingResource()
	var _ resource.Resource = r
}

func TestFirewallDefaultsResource_Schema(t *testing.T) {
	r := NewFirewallDefaultsResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages OpenWrt firewall default policies." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "input", "output", "forward", "synflood_protect", "synflood_rate", "drop_invalid", "auto_helper"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestFirewallDefaultsResource_Metadata(t *testing.T) {
	r := NewFirewallDefaultsResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_firewall_defaults" {
		t.Fatalf("expected type name 'openwrt_firewall_defaults', got %q", resp.TypeName)
	}
}

func TestFirewallDefaultsResource_ImplementsResource(t *testing.T) {
	r := NewFirewallDefaultsResource()
	var _ resource.Resource = r
}
