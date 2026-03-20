package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewSysNetDevicesDataSource() datasource.DataSource {
	return &sysNetDevicesDataSource{}
}

type sysNetDevicesDataSource struct {
	client *JsonRpcClient
}

type sysNetDeviceModel struct {
	Name      types.String `tfsdk:"name"`
	MTU       types.Int64  `tfsdk:"mtu"`
	MACAddr   types.String `tfsdk:"macaddr"`
	IPAddr    types.String `tfsdk:"ipaddr"`
	IP6Addr   types.String `tfsdk:"ip6addr"`
	RXBytes   types.Int64  `tfsdk:"rx_bytes"`
	TXBytes   types.Int64  `tfsdk:"tx_bytes"`
	RXPackets types.Int64  `tfsdk:"rx_packets"`
	TXPackets types.Int64  `tfsdk:"tx_packets"`
	Flags     types.String `tfsdk:"flags"`
}

type sysNetDevicesModel struct {
	ID      types.String `tfsdk:"id"`
	Devices types.List   `tfsdk:"devices"`
}

func (d *sysNetDevicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sys_net_devices"
}

func (d *sysNetDevicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about network devices from OpenWrt.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:    true,
				Description: "Internal ID.",
			},
			"devices": dsschema.ListAttribute{
				Computed:    true,
				Description: "List of network devices.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":       types.StringType,
						"mtu":        types.Int64Type,
						"macaddr":    types.StringType,
						"ipaddr":     types.StringType,
						"ip6addr":    types.StringType,
						"rx_bytes":   types.Int64Type,
						"tx_bytes":   types.Int64Type,
						"rx_packets": types.Int64Type,
						"tx_packets": types.Int64Type,
						"flags":      types.StringType,
					},
				},
			},
		},
	}
}

func (d *sysNetDevicesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*JsonRpcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			"Expected *JsonRpcClient",
		)
		return
	}
	d.client = client
}

func (d *sysNetDevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider must be configured before using this data source.")
		return
	}

	raw, err := d.client.SysCall(ctx, "net.deviceinfo")
	if err != nil {
		resp.Diagnostics.AddError("Error calling /rpc/sys", err.Error())
		return
	}

	var deviceInfo []map[string]interface{}
	if err := json.Unmarshal(raw, &deviceInfo); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	devices := make([]sysNetDeviceModel, 0, len(deviceInfo))
	for _, dev := range deviceInfo {
		dm := sysNetDeviceModel{}
		if v, ok := dev["name"].(string); ok {
			dm.Name = types.StringValue(v)
		}
		if v, ok := dev["mtu"].(float64); ok {
			dm.MTU = types.Int64Value(int64(v))
		}
		if v, ok := dev["macaddr"].(string); ok {
			dm.MACAddr = types.StringValue(v)
		}
		if v, ok := dev["ipaddr"].(string); ok {
			dm.IPAddr = types.StringValue(v)
		}
		if v, ok := dev["ip6addr"].(string); ok {
			dm.IP6Addr = types.StringValue(v)
		}
		if v, ok := dev["rx_bytes"].(float64); ok {
			dm.RXBytes = types.Int64Value(int64(v))
		}
		if v, ok := dev["tx_bytes"].(float64); ok {
			dm.TXBytes = types.Int64Value(int64(v))
		}
		if v, ok := dev["rx_packets"].(float64); ok {
			dm.RXPackets = types.Int64Value(int64(v))
		}
		if v, ok := dev["tx_packets"].(float64); ok {
			dm.TXPackets = types.Int64Value(int64(v))
		}
		if v, ok := dev["flags"].(string); ok {
			dm.Flags = types.StringValue(v)
		}
		devices = append(devices, dm)
	}

	devicesList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":       types.StringType,
			"mtu":        types.Int64Type,
			"macaddr":    types.StringType,
			"ipaddr":     types.StringType,
			"ip6addr":    types.StringType,
			"rx_bytes":   types.Int64Type,
			"tx_bytes":   types.Int64Type,
			"rx_packets": types.Int64Type,
			"tx_packets": types.Int64Type,
			"flags":      types.StringType,
		},
	}, devices)
	resp.Diagnostics.Append(diags...)

	state := sysNetDevicesModel{
		ID:      types.StringValue("net.deviceinfo"),
		Devices: devicesList,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
