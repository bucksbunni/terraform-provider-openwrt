package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestSystemResource_Schema(t *testing.T) {
	r := NewSystemResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages OpenWrt system settings." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "hostname", "ttylogin", "log_size", "urandom_seed", "zonename", "log_proto", "conloglevel", "cronloglevel"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestSystemResource_Metadata(t *testing.T) {
	r := NewSystemResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_system" {
		t.Fatalf("expected type name 'openwrt_system', got %q", resp.TypeName)
	}
}

func TestSystemResource_ImplementsResource(t *testing.T) {
	r := NewSystemResource()
	var _ resource.Resource = r
}

func TestSystemNTPResource_Schema(t *testing.T) {
	r := NewSystemNTPResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages OpenWrt NTP (timeserver) configuration." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "name", "enabled", "server", "port"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestSystemNTPResource_Metadata(t *testing.T) {
	r := NewSystemNTPResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_system_ntp" {
		t.Fatalf("expected type name 'openwrt_system_ntp', got %q", resp.TypeName)
	}
}

func TestSystemNTPResource_ImplementsResource(t *testing.T) {
	r := NewSystemNTPResource()
	var _ resource.Resource = r
}

func TestSystemLEDResource_Schema(t *testing.T) {
	r := NewSystemLEDResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages OpenWrt system LED configuration." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "name", "sysfs", "trigger", "mode", "dev", "default"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestSystemLEDResource_Metadata(t *testing.T) {
	r := NewSystemLEDResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_system_led" {
		t.Fatalf("expected type name 'openwrt_system_led', got %q", resp.TypeName)
	}
}

func TestSystemLEDResource_ImplementsResource(t *testing.T) {
	r := NewSystemLEDResource()
	var _ resource.Resource = r
}

func TestDropbearResource_Schema(t *testing.T) {
	r := NewDropbearResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages OpenWrt Dropbear SSH server settings." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "name", "password_auth", "port", "root_password_auth", "root_login"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestDropbearResource_Metadata(t *testing.T) {
	r := NewDropbearResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_dropbear" {
		t.Fatalf("expected type name 'openwrt_dropbear', got %q", resp.TypeName)
	}
}

func TestDropbearResource_ImplementsResource(t *testing.T) {
	r := NewDropbearResource()
	var _ resource.Resource = r
}
