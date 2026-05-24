package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestWirelessDeviceResource_Schema(t *testing.T) {
	r := NewWirelessDeviceResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages an OpenWrt wireless radio device." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "name", "type", "path", "band", "channel", "htmode", "country", "disabled"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestWirelessDeviceResource_Metadata(t *testing.T) {
	r := NewWirelessDeviceResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_wireless_device" {
		t.Fatalf("expected type name 'openwrt_wireless_device', got %q", resp.TypeName)
	}
}

func TestWirelessDeviceResource_ImplementsResource(t *testing.T) {
	r := NewWirelessDeviceResource()
	var _ resource.Resource = r
}

func TestWirelessInterfaceResource_Schema(t *testing.T) {
	r := NewWirelessInterfaceResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	s := resp.Schema

	if s.Description != "Manages an OpenWrt wireless interface (WiFi SSID)." {
		t.Fatalf("expected correct description, got %q", s.Description)
	}

	expectedAttrs := []string{"id", "name", "device", "mode", "ssid", "encryption", "key", "network", "disabled", "hidden", "macfilter", "maclist", "isolate"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Fatalf("expected %s attribute", attr)
		}
	}
}

func TestWirelessInterfaceResource_Metadata(t *testing.T) {
	r := NewWirelessInterfaceResource()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "openwrt",
	}, &resp)

	if resp.TypeName != "openwrt_wireless_iface" {
		t.Fatalf("expected type name 'openwrt_wireless_iface', got %q", resp.TypeName)
	}
}

func TestWirelessInterfaceResource_ImplementsResource(t *testing.T) {
	r := NewWirelessInterfaceResource()
	var _ resource.Resource = r
}
