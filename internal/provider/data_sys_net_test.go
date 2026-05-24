package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestSysNetDevicesDataSource_Schema(t *testing.T) {
	d := NewSysNetDevicesDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	s := resp.Schema

	expectedAttrs := []string{"id", "devices"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestSysNetDevicesDataSource_Metadata(t *testing.T) {
	d := NewSysNetDevicesDataSource()
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_sys_net_devices" {
		t.Fatalf("expected type name 'openwrt_sys_net_devices', got %q", resp.TypeName)
	}
}

func TestSysNetDevicesDataSource_ImplementsDataSource(t *testing.T) {
	d := NewSysNetDevicesDataSource()
	var _ datasource.DataSource = d
}

func TestSysNetRoutesDataSource_Schema(t *testing.T) {
	d := NewSysNetRoutesDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	s := resp.Schema

	expectedAttrs := []string{"id", "routes"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestSysNetRoutesDataSource_Metadata(t *testing.T) {
	d := NewSysNetRoutesDataSource()
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_sys_net_routes" {
		t.Fatalf("expected type name 'openwrt_sys_net_routes', got %q", resp.TypeName)
	}
}

func TestSysNetRoutesDataSource_ImplementsDataSource(t *testing.T) {
	d := NewSysNetRoutesDataSource()
	var _ datasource.DataSource = d
}

func TestSysNetRoutes6DataSource_Schema(t *testing.T) {
	d := NewSysNetRoutes6DataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	s := resp.Schema

	expectedAttrs := []string{"id", "routes"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestSysNetRoutes6DataSource_Metadata(t *testing.T) {
	d := NewSysNetRoutes6DataSource()
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_sys_net_routes6" {
		t.Fatalf("expected type name 'openwrt_sys_net_routes6', got %q", resp.TypeName)
	}
}

func TestSysNetRoutes6DataSource_ImplementsDataSource(t *testing.T) {
	d := NewSysNetRoutes6DataSource()
	var _ datasource.DataSource = d
}

func TestSysNetArpTableDataSource_Schema(t *testing.T) {
	d := NewSysNetArpTableDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	s := resp.Schema

	expectedAttrs := []string{"id", "entries"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestSysNetArpTableDataSource_Metadata(t *testing.T) {
	d := NewSysNetArpTableDataSource()
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_sys_net_arptable" {
		t.Fatalf("expected type name 'openwrt_sys_net_arptable', got %q", resp.TypeName)
	}
}

func TestSysNetArpTableDataSource_ImplementsDataSource(t *testing.T) {
	d := NewSysNetArpTableDataSource()
	var _ datasource.DataSource = d
}

func TestSysNetConntrackDataSource_Schema(t *testing.T) {
	d := NewSysNetConntrackDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	s := resp.Schema

	expectedAttrs := []string{"id", "entries"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestSysNetConntrackDataSource_Metadata(t *testing.T) {
	d := NewSysNetConntrackDataSource()
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_sys_net_conntrack" {
		t.Fatalf("expected type name 'openwrt_sys_net_conntrack', got %q", resp.TypeName)
	}
}

func TestSysNetConntrackDataSource_ImplementsDataSource(t *testing.T) {
	d := NewSysNetConntrackDataSource()
	var _ datasource.DataSource = d
}

func TestSysProcessListDataSource_Schema(t *testing.T) {
	d := NewSysProcessListDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	s := resp.Schema

	expectedAttrs := []string{"id", "processes"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestSysProcessListDataSource_Metadata(t *testing.T) {
	d := NewSysProcessListDataSource()
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_sys_process_list" {
		t.Fatalf("expected type name 'openwrt_sys_process_list', got %q", resp.TypeName)
	}
}

func TestSysProcessListDataSource_ImplementsDataSource(t *testing.T) {
	d := NewSysProcessListDataSource()
	var _ datasource.DataSource = d
}

func TestSysWirelessDataSource_Schema(t *testing.T) {
	d := NewSysWirelessDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	s := resp.Schema

	expectedAttrs := []string{"id", "ifname", "channel", "frequency", "txpower", "signal", "noise", "bitrate", "ssid", "mode", "essid", "encryption"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestSysWirelessDataSource_Metadata(t *testing.T) {
	d := NewSysWirelessDataSource()
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_sys_wireless" {
		t.Fatalf("expected type name 'openwrt_sys_wireless', got %q", resp.TypeName)
	}
}

func TestSysWirelessDataSource_ImplementsDataSource(t *testing.T) {
	d := NewSysWirelessDataSource()
	var _ datasource.DataSource = d
}
