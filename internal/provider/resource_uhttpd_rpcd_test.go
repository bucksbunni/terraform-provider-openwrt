package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestUHTTPdResource_Schema(t *testing.T) {
	r := NewUHTTPdResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages OpenWrt uHTTPd web server settings." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "name", "listen_http", "listen_https", "redirect_https", "home", "rfc1918_filter", "max_requests", "max_connections", "cert", "key", "cgi_prefix", "lua_prefix", "script_timeout", "network_timeout", "http_keepalive", "tcp_keepalive", "ubus_prefix"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestUHTTPdResource_Metadata(t *testing.T) {
	r := NewUHTTPdResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_uhttpd" {
		t.Fatalf("expected type name 'openwrt_uhttpd', got %q", resp.TypeName)
	}
}

func TestUHTTPdResource_ImplementsResource(t *testing.T) {
	r := NewUHTTPdResource()
	var _ resource.Resource = r
}

func TestUHTTPdCertResource_Schema(t *testing.T) {
	r := NewUHTTPdCertResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages OpenWrt uHTTPd certificate generation settings." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "name", "days", "key_type", "bits", "ec_curve", "country", "state", "location", "commonname"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestUHTTPdCertResource_Metadata(t *testing.T) {
	r := NewUHTTPdCertResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_uhttpd_cert" {
		t.Fatalf("expected type name 'openwrt_uhttpd_cert', got %q", resp.TypeName)
	}
}

func TestUHTTPdCertResource_ImplementsResource(t *testing.T) {
	r := NewUHTTPdCertResource()
	var _ resource.Resource = r
}

func TestRPCDResource_Schema(t *testing.T) {
	r := NewRPCDResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages OpenWrt RPC daemon (ubus) settings." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "socket", "timeout"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestRPCDResource_Metadata(t *testing.T) {
	r := NewRPCDResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_rpcd" {
		t.Fatalf("expected type name 'openwrt_rpcd', got %q", resp.TypeName)
	}
}

func TestRPCDResource_ImplementsResource(t *testing.T) {
	r := NewRPCDResource()
	var _ resource.Resource = r
}
