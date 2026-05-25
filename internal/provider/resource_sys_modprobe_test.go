package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestSysModprobeResource_Schema(t *testing.T) {
	r := NewSysModprobeResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Loads or unloads a kernel module on the OpenWrt device." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "name", "action", "param", "output"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestSysModprobeResource_Metadata(t *testing.T) {
	r := NewSysModprobeResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_sys_modprobe" {
		t.Fatalf("expected type name 'openwrt_sys_modprobe', got %q", resp.TypeName)
	}
}

func TestSysModprobeResource_ImplementsResource(t *testing.T) {
	r := NewSysModprobeResource()
	var _ resource.Resource = r
}