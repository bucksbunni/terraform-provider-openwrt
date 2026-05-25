package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestSysRebootResource_Schema(t *testing.T) {
	r := NewSysRebootResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Triggers a reboot of the OpenWrt device." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "delay"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestSysRebootResource_Metadata(t *testing.T) {
	r := NewSysRebootResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_sys_reboot" {
		t.Fatalf("expected type name 'openwrt_sys_reboot', got %q", resp.TypeName)
	}
}

func TestSysRebootResource_ImplementsResource(t *testing.T) {
	r := NewSysRebootResource()
	var _ resource.Resource = r
}