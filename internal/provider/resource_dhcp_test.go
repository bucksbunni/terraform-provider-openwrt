package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestDHCPPoolResource_Schema(t *testing.T) {
	r := NewDHCPPoolResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages a DHCP pool for an OpenWrt network interface." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "name", "interface", "start", "limit", "leasetime", "dhcpv4", "dhcpv6", "ra", "ra_flags", "ignore"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestDHCPPoolResource_Metadata(t *testing.T) {
	r := NewDHCPPoolResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_dhcp_pool" {
		t.Fatalf("expected type name 'openwrt_dhcp_pool', got %q", resp.TypeName)
	}
}

func TestDHCPPoolResource_ImplementsResource(t *testing.T) {
	r := NewDHCPPoolResource()
	var _ resource.Resource = r
}

func TestDHCPDNSMasqResource_Schema(t *testing.T) {
	r := NewDHCPDNSMasqResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages DNS/DHCP settings (dnsmasq) for OpenWrt." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "domainneeded", "localise_queries", "rebind_protection", "rebind_localhost", "local", "domain", "expand_hosts", "cachesize", "authoritative", "readethers", "leasefile", "resolvfile", "localservice", "ednspacket_max", "confdir", "rebind_domain", "server"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestDHCPDNSMasqResource_Metadata(t *testing.T) {
	r := NewDHCPDNSMasqResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_dhcp_dnsmasq" {
		t.Fatalf("expected type name 'openwrt_dhcp_dnsmasq', got %q", resp.TypeName)
	}
}

func TestDHCPDNSMasqResource_ImplementsResource(t *testing.T) {
	r := NewDHCPDNSMasqResource()
	var _ resource.Resource = r
}

func TestDHCPOdhcpdResource_Schema(t *testing.T) {
	r := NewDHCPOdhcpdResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages OpenWrt DHCPv6 (odhcpd) settings." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "maindhcp", "leasefile", "leasetrigger", "loglevel"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestDHCPOdhcpdResource_Metadata(t *testing.T) {
	r := NewDHCPOdhcpdResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_dhcp_odhcpd" {
		t.Fatalf("expected type name 'openwrt_dhcp_odhcpd', got %q", resp.TypeName)
	}
}

func TestDHCPOdhcpdResource_ImplementsResource(t *testing.T) {
	r := NewDHCPOdhcpdResource()
	var _ resource.Resource = r
}

func TestDHCPHostResource_Schema(t *testing.T) {
	r := NewDHCPHostResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages a static DHCP host reservation in OpenWrt." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "name", "ip", "mac", "leasetime"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestDHCPHostResource_Metadata(t *testing.T) {
	r := NewDHCPHostResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_dhcp_host" {
		t.Fatalf("expected type name 'openwrt_dhcp_host', got %q", resp.TypeName)
	}
}

func TestDHCPHostResource_ImplementsResource(t *testing.T) {
	r := NewDHCPHostResource()
	var _ resource.Resource = r
}
