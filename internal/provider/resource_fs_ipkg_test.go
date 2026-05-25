package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestFSFileResource_Schema(t *testing.T) {
	r := NewFSFileResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages a file on the OpenWrt filesystem via LuCI /rpc/fs." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "path", "content"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestFSFileResource_Metadata(t *testing.T) {
	r := NewFSFileResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_fs_file" {
		t.Fatalf("expected type name 'openwrt_fs_file', got %q", resp.TypeName)
	}
}

func TestFSFileResource_ImplementsResource(t *testing.T) {
	r := NewFSFileResource()
	var _ resource.Resource = r
}

func TestFSFileResource_ImplementsImportState(t *testing.T) {
	r := NewFSFileResource()
	_, ok := r.(resource.ResourceWithImportState)
	if !ok {
		t.Error("FSFileResource should implement ResourceWithImportState")
	}
}

func TestIPKGPackageResource_Schema(t *testing.T) {
	r := NewIPKGPackageResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages an OpenWrt package via LuCI /rpc/ipkg." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "name", "autoremove"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestIPKGPackageResource_Metadata(t *testing.T) {
	r := NewIPKGPackageResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_ipkg_package" {
		t.Fatalf("expected type name 'openwrt_ipkg_package', got %q", resp.TypeName)
	}
}

func TestIPKGPackageResource_ImplementsResource(t *testing.T) {
	r := NewIPKGPackageResource()
	var _ resource.Resource = r
}

func TestIPKGPackageResource_ImplementsImportState(t *testing.T) {
	r := NewIPKGPackageResource()
	_, ok := r.(resource.ResourceWithImportState)
	if !ok {
		t.Error("IPKGPackageResource should implement ResourceWithImportState")
	}
}
