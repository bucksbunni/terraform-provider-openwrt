package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestSysRPCDataSource_Schema(t *testing.T) {
	d := NewSysRPCDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	s := resp.Schema

	expectedAttrs := []string{"id", "method", "params_json", "result_json"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestSysRPCDataSource_Metadata(t *testing.T) {
	d := NewSysRPCDataSource()
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_sys_rpc" {
		t.Fatalf("expected type name 'openwrt_sys_rpc', got %q", resp.TypeName)
	}
}

func TestSysRPCDataSource_ImplementsDataSource(t *testing.T) {
	d := NewSysRPCDataSource()
	var _ datasource.DataSource = d
}

func TestSysHostnameDataSource_Schema(t *testing.T) {
	d := NewSysHostnameDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	s := resp.Schema

	expectedAttrs := []string{"id", "hostname"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestSysHostnameDataSource_Metadata(t *testing.T) {
	d := NewSysHostnameDataSource()
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_sys_hostname" {
		t.Fatalf("expected type name 'openwrt_sys_hostname', got %q", resp.TypeName)
	}
}

func TestSysHostnameDataSource_ImplementsDataSource(t *testing.T) {
	d := NewSysHostnameDataSource()
	var _ datasource.DataSource = d
}

func TestSysUptimeDataSource_Schema(t *testing.T) {
	d := NewSysUptimeDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	s := resp.Schema

	expectedAttrs := []string{"id", "uptime"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestSysUptimeDataSource_Metadata(t *testing.T) {
	d := NewSysUptimeDataSource()
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_sys_uptime" {
		t.Fatalf("expected type name 'openwrt_sys_uptime', got %q", resp.TypeName)
	}
}

func TestSysUptimeDataSource_ImplementsDataSource(t *testing.T) {
	d := NewSysUptimeDataSource()
	var _ datasource.DataSource = d
}

func TestSysInitDataSource_Schema(t *testing.T) {
	d := NewSysInitDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	s := resp.Schema

	expectedAttrs := []string{"id", "scripts"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestSysInitDataSource_Metadata(t *testing.T) {
	d := NewSysInitDataSource()
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_sys_init" {
		t.Fatalf("expected type name 'openwrt_sys_init', got %q", resp.TypeName)
	}
}

func TestSysInitDataSource_ImplementsDataSource(t *testing.T) {
	d := NewSysInitDataSource()
	var _ datasource.DataSource = d
}
