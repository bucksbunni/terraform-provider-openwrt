package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewSysNetRoutesDataSource() datasource.DataSource {
	return &sysNetRoutesDataSource{}
}

type sysNetRoutesDataSource struct {
	client *JsonRpcClient
}

type sysNetRouteModel struct {
	Dest     types.String `tfsdk:"dest"`
	Gateway  types.String `tfsdk:"gateway"`
	Metric   types.Int64  `tfsdk:"metric"`
	RefCount types.Int64  `tfsdk:"refcount"`
	UseCount types.Int64  `tfsdk:"usecount"`
	IRTT     types.Int64  `tfsdk:"irtt"`
	Flags    types.String `tfsdk:"flags"`
	Device   types.String `tfsdk:"device"`
}

type sysNetRoutesModel struct {
	ID     types.String `tfsdk:"id"`
	Routes types.List   `tfsdk:"routes"`
}

func (d *sysNetRoutesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sys_net_routes"
}

func (d *sysNetRoutesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves the IPv4 routing table from OpenWrt.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:    true,
				Description: "Internal ID.",
			},
			"routes": dsschema.ListAttribute{
				Computed:    true,
				Description: "List of IPv4 routes.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"dest":     types.StringType,
						"gateway":  types.StringType,
						"metric":   types.Int64Type,
						"refcount": types.Int64Type,
						"usecount": types.Int64Type,
						"irtt":     types.Int64Type,
						"flags":    types.StringType,
						"device":   types.StringType,
					},
				},
			},
		},
	}
}

func (d *sysNetRoutesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sysNetRoutesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider must be configured before using this data source.")
		return
	}

	// luci.sys.net.routes() was removed in modern LuCI, so read the kernel
	// routing table directly. /proc/net/route columns:
	//   Iface Destination Gateway Flags RefCnt Use Metric Mask MTU Window IRTT
	// Destination/Gateway are little-endian hex IPv4 words.
	out, err := d.client.SysExec(ctx, "cat /proc/net/route")
	if err != nil {
		resp.Diagnostics.AddError("Error calling /rpc/sys", err.Error())
		return
	}

	lines := nonHeaderLines(out, 1)
	routes := make([]sysNetRouteModel, 0, len(lines))
	for _, line := range lines {
		f := strings.Fields(line)
		if len(f) < 11 {
			continue
		}
		routes = append(routes, sysNetRouteModel{
			Dest:     types.StringValue(hexLEToIPv4(f[1])),
			Gateway:  types.StringValue(hexLEToIPv4(f[2])),
			Flags:    types.StringValue(f[3]),
			RefCount: types.Int64Value(atoiSafe(f[4])),
			UseCount: types.Int64Value(atoiSafe(f[5])),
			Metric:   types.Int64Value(atoiSafe(f[6])),
			IRTT:     types.Int64Value(atoiSafe(f[10])),
			Device:   types.StringValue(f[0]),
		})
	}

	routesList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"dest":     types.StringType,
			"gateway":  types.StringType,
			"metric":   types.Int64Type,
			"refcount": types.Int64Type,
			"usecount": types.Int64Type,
			"irtt":     types.Int64Type,
			"flags":    types.StringType,
			"device":   types.StringType,
		},
	}, routes)
	resp.Diagnostics.Append(diags...)

	state := sysNetRoutesModel{
		ID:     types.StringValue("net.routes"),
		Routes: routesList,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
