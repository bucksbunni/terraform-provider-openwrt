package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestProviderMetadata(t *testing.T) {
	newProv := NewProvider("test-version")
	p := newProv().(*openwrtProvider)

	var resp provider.MetadataResponse
	p.Metadata(context.Background(), provider.MetadataRequest{}, &resp)

	if resp.TypeName != "openwrt" {
		t.Fatalf("expected TypeName 'openwrt', got %q", resp.TypeName)
	}
	if resp.Version != "test-version" {
		t.Fatalf("expected Version 'test-version', got %q", resp.Version)
	}
}

func TestProviderSchemaHasExpectedAttributes(t *testing.T) {
	newProv := NewProvider("test-version")
	p := newProv().(*openwrtProvider)

	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)

	s := resp.Schema

	if _, ok := s.Attributes["host"]; !ok {
		t.Fatal("expected provider schema to have 'host' attribute")
	}
	if _, ok := s.Attributes["username"]; !ok {
		t.Fatal("expected provider schema to have 'username' attribute")
	}
	if _, ok := s.Attributes["password"]; !ok {
		t.Fatal("expected provider schema to have 'password' attribute")
	}
	if _, ok := s.Attributes["insecure"]; !ok {
		t.Fatal("expected provider schema to have 'insecure' attribute")
	}

	if !s.Attributes["host"].IsRequired() {
		t.Fatal("expected 'host' to be required")
	}
	if !s.Attributes["username"].IsRequired() {
		t.Fatal("expected 'username' to be required")
	}
	if !s.Attributes["password"].IsRequired() {
		t.Fatal("expected 'password' to be required")
	}
	if s.Attributes["insecure"].IsRequired() {
		t.Fatal("expected 'insecure' to be optional")
	}
}

func TestProviderResources(t *testing.T) {
	newProv := NewProvider("test-version")
	p := newProv()

	resources := p.Resources(context.Background())

	if len(resources) == 0 {
		t.Fatal("expected at least one resource")
	}
}

func TestProviderDataSources(t *testing.T) {
	newProv := NewProvider("test-version")
	p := newProv()

	dataSources := p.DataSources(context.Background())

	if len(dataSources) == 0 {
		t.Fatal("expected at least one data source")
	}
}
