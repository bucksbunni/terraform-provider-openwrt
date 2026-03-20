package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewSysWirelessDataSource() datasource.DataSource {
	return &sysWirelessDataSource{}
}

type sysWirelessDataSource struct {
	client *JsonRpcClient
}

type sysWirelessModel struct {
	ID         types.String `tfsdk:"id"`
	Ifname     types.String `tfsdk:"ifname"`
	Channel    types.Int64  `tfsdk:"channel"`
	Frequency  types.Int64  `tfsdk:"frequency"`
	TxPower    types.Int64  `tfsdk:"txpower"`
	Signal     types.Int64  `tfsdk:"signal"`
	Noise      types.Int64  `tfsdk:"noise"`
	Bitrate    types.Int64  `tfsdk:"bitrate"`
	SSID       types.String `tfsdk:"ssid"`
	Mode       types.String `tfsdk:"mode"`
	ESSID      types.String `tfsdk:"essid"`
	Encryption types.String `tfsdk:"encryption"`
}

type sysWirelessInfoModel struct {
	ID      types.String `tfsdk:"id"`
	PhyName types.String `tfsdk:"phy_name"`
	IFName  types.String `tfsdk:"ifname"`
}

func NewSysWirelessInfoDataSource() datasource.DataSource {
	return &sysWirelessInfoDataSource{}
}

type sysWirelessInfoDataSource struct {
	client *JsonRpcClient
}

func (d *sysWirelessDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sys_wireless"
}

func (d *sysWirelessDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves wireless information for a given interface from OpenWrt.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:    true,
				Description: "Internal ID.",
			},
			"ifname": dsschema.StringAttribute{
				Required:    true,
				Description: "Wireless interface name (e.g., 'wlan0').",
			},
			"channel": dsschema.Int64Attribute{
				Computed:    true,
				Description: "Current channel.",
			},
			"frequency": dsschema.Int64Attribute{
				Computed:    true,
				Description: "Frequency in MHz.",
			},
			"txpower": dsschema.Int64Attribute{
				Computed:    true,
				Description: "Transmit power in dBm.",
			},
			"signal": dsschema.Int64Attribute{
				Computed:    true,
				Description: "Signal strength in dBm.",
			},
			"noise": dsschema.Int64Attribute{
				Computed:    true,
				Description: "Noise level in dBm.",
			},
			"bitrate": dsschema.Int64Attribute{
				Computed:    true,
				Description: "Bitrate in Mb/s.",
			},
			"ssid": dsschema.StringAttribute{
				Computed:    true,
				Description: "SSID name.",
			},
			"mode": dsschema.StringAttribute{
				Computed:    true,
				Description: "Operation mode.",
			},
			"essid": dsschema.StringAttribute{
				Computed:    true,
				Description: "Extended SSID.",
			},
			"encryption": dsschema.StringAttribute{
				Computed:    true,
				Description: "Encryption mode.",
			},
		},
	}
}

func (d *sysWirelessDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sysWirelessDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider must be configured before using this data source.")
		return
	}

	var config sysWirelessModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ifname := config.Ifname.ValueString()
	if ifname == "" {
		resp.Diagnostics.AddError("Missing ifname", "ifname must be set")
		return
	}

	raw, err := d.client.SysCall(ctx, "wifi.getiwinfo", ifname)
	if err != nil {
		resp.Diagnostics.AddError("Error calling /rpc/sys", err.Error())
		return
	}

	var info map[string]interface{}
	if err := json.Unmarshal(raw, &info); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	state := sysWirelessModel{
		ID:     types.StringValue("wifi." + ifname),
		Ifname: config.Ifname,
	}

	if v, ok := info["channel"].(float64); ok {
		state.Channel = types.Int64Value(int64(v))
	}
	if v, ok := info["frequency"].(float64); ok {
		state.Frequency = types.Int64Value(int64(v))
	}
	if v, ok := info["txpower"].(float64); ok {
		state.TxPower = types.Int64Value(int64(v))
	}
	if v, ok := info["signal"].(float64); ok {
		state.Signal = types.Int64Value(int64(v))
	}
	if v, ok := info["noise"].(float64); ok {
		state.Noise = types.Int64Value(int64(v))
	}
	if v, ok := info["bitrate"].(float64); ok {
		state.Bitrate = types.Int64Value(int64(v))
	}
	if v, ok := info["ssid"].(string); ok {
		state.SSID = types.StringValue(v)
	}
	if v, ok := info["mode"].(string); ok {
		state.Mode = types.StringValue(v)
	}
	if v, ok := info["essid"].(string); ok {
		state.ESSID = types.StringValue(v)
	}
	if v, ok := info["encryption"].(string); ok {
		state.Encryption = types.StringValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *sysWirelessInfoDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sys_wireless_info"
}

func (d *sysWirelessInfoDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves available wireless interfaces from OpenWrt.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:    true,
				Description: "Internal ID.",
			},
			"interfaces": dsschema.ListAttribute{
				Computed:    true,
				Description: "List of wireless interfaces.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":       types.StringType,
						"phy_name": types.StringType,
						"ifname":   types.StringType,
					},
				},
			},
		},
	}
}

func (d *sysWirelessInfoDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sysWirelessInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider must be configured before using this data source.")
		return
	}

	raw, err := d.client.SysCall(ctx, "wifi.getiwinfo", "")
	if err != nil {
		resp.Diagnostics.AddError("Error calling /rpc/sys", err.Error())
		return
	}

	var info map[string]interface{}
	if err := json.Unmarshal(raw, &info); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	interfaces := make([]sysWirelessInfoModel, 0, 1)
	iface := sysWirelessInfoModel{
		ID: types.StringValue("wifi"),
	}
	if v, ok := info["phyname"].(string); ok {
		iface.PhyName = types.StringValue(v)
	}
	if v, ok := info["ifname"].(string); ok {
		iface.IFName = types.StringValue(v)
	}
	interfaces = append(interfaces, iface)

	interfacesList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":       types.StringType,
			"phy_name": types.StringType,
			"ifname":   types.StringType,
		},
	}, interfaces)
	resp.Diagnostics.Append(diags...)

	state := struct {
		ID         types.String `tfsdk:"id"`
		Interfaces types.List   `tfsdk:"interfaces"`
	}{
		ID:         types.StringValue("wifi"),
		Interfaces: interfacesList,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
