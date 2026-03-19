package resource

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestUCISectionResource_Schema(t *testing.T) {
	r := NewUCISectionResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages an OpenWrt UCI section and its key/value options." {
		t.Fatalf("expected correct description")
	}

	if _, ok := s.Attributes["id"]; !ok {
		t.Fatal("expected id attribute")
	}
	if _, ok := s.Attributes["config"]; !ok {
		t.Fatal("expected config attribute")
	}
	if _, ok := s.Attributes["type"]; !ok {
		t.Fatal("expected type attribute")
	}
	if _, ok := s.Attributes["name"]; !ok {
		t.Fatal("expected name attribute")
	}
	if _, ok := s.Attributes["options"]; !ok {
		t.Fatal("expected options attribute")
	}
}

func TestUCISectionResource_Metadata(t *testing.T) {
	r := NewUCISectionResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_uci_section" {
		t.Fatalf("expected type name 'openwrt_uci_section', got %q", resp.TypeName)
	}
}

func TestUCISectionResource_ImplementsResource(t *testing.T) {
	r := NewUCISectionResource()
	var _ resource.Resource = r
}
