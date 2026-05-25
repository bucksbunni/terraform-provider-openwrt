package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewSysWirelessIfacesDataSource() datasource.DataSource {
	return &sysWirelessIfacesDataSource{}
}

type sysWirelessIfacesDataSource struct {
	client *JsonRpcClient
}

type sysWirelessIfaceModel struct {
	Name       types.String   `tfsdk:"name"`
	Device     types.String   `tfsdk:"device"`
	Mode       types.String   `tfsdk:"mode"`
	SSID       types.String   `tfsdk:"ssid"`
	Encryption types.String   `tfsdk:"encryption"`
	Network    types.List     `tfsdk:"network"`
	Disabled   types.Bool     `tfsdk:"disabled"`
	Hidden     types.Bool     `tfsdk:"hidden"`
	MACFilter  types.String   `tfsdk:"macfilter"`
	MACList    types.List     `tfsdk:"maclist"`
	Isolate    types.Bool     `tfsdk:"isolate"`
}

type sysWirelessIfacesModel struct {
	ID       types.String `tfsdk:"id"`
	Ifaces   types.List   `tfsdk:"interfaces"`
}

func (d *sysWirelessIfacesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sys_wireless_ifaces"
}

func (d *sysWirelessIfacesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Discovers wireless interfaces from OpenWrt.",
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
						"name":       types.StringType,
						"device":     types.StringType,
						"mode":       types.StringType,
						"ssid":       types.StringType,
						"encryption": types.StringType,
						"network":    types.ListType{ElemType: types.StringType},
						"disabled":   types.BoolType,
						"hidden":     types.BoolType,
						"macfilter":  types.StringType,
						"maclist":    types.ListType{ElemType: types.StringType},
						"isolate":    types.BoolType,
					},
				},
			},
		},
	}
}

func (d *sysWirelessIfacesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sysWirelessIfacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider must be configured before using this data source.")
		return
	}

	ifaces, err := d.client.UCIForeach(ctx, "wireless", "wifi-iface")
	if err != nil {
		resp.Diagnostics.AddError("Error reading wireless interfaces", err.Error())
		return
	}

	ifaceList := make([]sysWirelessIfaceModel, 0, len(ifaces))
	for _, iface := range ifaces {
		i := sysWirelessIfaceModel{}

		if v, ok := iface[".name"].(string); ok {
			i.Name = types.StringValue(v)
		}
		if v, ok := iface["device"].(string); ok {
			i.Device = types.StringValue(v)
		}
		if v, ok := iface["mode"].(string); ok {
			i.Mode = types.StringValue(v)
		}
		if v, ok := iface["ssid"].(string); ok {
			i.SSID = types.StringValue(v)
		}
		if v, ok := iface["encryption"].(string); ok {
			i.Encryption = types.StringValue(v)
		}
		if v, ok := iface["network"].(string); ok && v != "" {
			networks := strings.Split(v, " ")
			i.Network, _ = types.ListValueFrom(ctx, types.StringType, networks)
		}
		if v, ok := iface["disabled"].(string); ok {
			i.Disabled = types.BoolValue(v == "1" || v == "true")
		}
		if v, ok := iface["hidden"].(string); ok {
			i.Hidden = types.BoolValue(v == "1" || v == "true")
		}
		if v, ok := iface["macfilter"].(string); ok {
			i.MACFilter = types.StringValue(v)
		}
		if v, ok := iface["maclist"].(string); ok && v != "" {
			maclist := strings.Split(v, " ")
			i.MACList, _ = types.ListValueFrom(ctx, types.StringType, maclist)
		}
		if v, ok := iface["isolate"].(string); ok {
			i.Isolate = types.BoolValue(v == "1" || v == "true")
		}

		ifaceList = append(ifaceList, i)
	}

	ifacesList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":       types.StringType,
			"device":     types.StringType,
			"mode":       types.StringType,
			"ssid":       types.StringType,
			"encryption": types.StringType,
			"network":    types.ListType{ElemType: types.StringType},
			"disabled":   types.BoolType,
			"hidden":     types.BoolType,
			"macfilter":  types.StringType,
			"maclist":    types.ListType{ElemType: types.StringType},
			"isolate":    types.BoolType,
		},
	}, ifaceList)
	resp.Diagnostics.Append(diags...)

	state := sysWirelessIfacesModel{
		ID:     types.StringValue("wireless_ifaces"),
		Ifaces: ifacesList,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}