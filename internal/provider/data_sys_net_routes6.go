package provider

import (
	"context"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewSysNetRoutes6DataSource() datasource.DataSource {
	return &sysNetRoutes6DataSource{}
}

type sysNetRoutes6DataSource struct {
	client *JsonRpcClient
}

type sysNetRoute6Model struct {
	Source   types.String `tfsdk:"source"`
	Dest     types.String `tfsdk:"dest"`
	Nexthop  types.String `tfsdk:"nexthop"`
	Metric   types.Int64  `tfsdk:"metric"`
	RefCount types.Int64  `tfsdk:"refcount"`
	UseCount types.Int64  `tfsdk:"usecount"`
	Flags    types.String `tfsdk:"flags"`
	Device   types.String `tfsdk:"device"`
}

type sysNetRoutes6Model struct {
	ID     types.String `tfsdk:"id"`
	Routes types.List   `tfsdk:"routes"`
}

func (d *sysNetRoutes6DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sys_net_routes6"
}

func (d *sysNetRoutes6DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves the IPv6 routing table from OpenWrt.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:    true,
				Description: "Internal ID.",
			},
			"routes": dsschema.ListAttribute{
				Computed:    true,
				Description: "List of IPv6 routes.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"source":   types.StringType,
						"dest":     types.StringType,
						"nexthop":  types.StringType,
						"metric":   types.Int64Type,
						"refcount": types.Int64Type,
						"usecount": types.Int64Type,
						"flags":    types.StringType,
						"device":   types.StringType,
					},
				},
			},
		},
	}
}

func (d *sysNetRoutes6DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sysNetRoutes6DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider must be configured before using this data source.")
		return
	}

	// luci.sys.net.routes6() was removed in modern LuCI, so read the kernel
	// IPv6 routing table directly. /proc/net/ipv6_route columns:
	//   dest_net dest_plen src_net src_plen next_hop metric refcnt use flags iface
	// All addresses/counters are hex; addresses are 32 hex characters.
	out, err := d.client.SysExec(ctx, "cat /proc/net/ipv6_route")
	if err != nil {
		resp.Diagnostics.AddError("Error calling /rpc/sys", err.Error())
		return
	}

	lines := nonHeaderLines(out, 0)
	routes := make([]sysNetRoute6Model, 0, len(lines))
	for _, line := range lines {
		f := strings.Fields(line)
		if len(f) < 10 {
			continue
		}
		routes = append(routes, sysNetRoute6Model{
			Source:   types.StringValue(hexToIPv6(f[2]) + "/" + strconv.FormatInt(hexToInt(f[3]), 10)),
			Dest:     types.StringValue(hexToIPv6(f[0]) + "/" + strconv.FormatInt(hexToInt(f[1]), 10)),
			Nexthop:  types.StringValue(hexToIPv6(f[4])),
			Metric:   types.Int64Value(hexToInt(f[5])),
			RefCount: types.Int64Value(hexToInt(f[6])),
			UseCount: types.Int64Value(hexToInt(f[7])),
			Flags:    types.StringValue(f[8]),
			Device:   types.StringValue(f[9]),
		})
	}

	routesList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"source":   types.StringType,
			"dest":     types.StringType,
			"nexthop":  types.StringType,
			"metric":   types.Int64Type,
			"refcount": types.Int64Type,
			"usecount": types.Int64Type,
			"flags":    types.StringType,
			"device":   types.StringType,
		},
	}, routes)
	resp.Diagnostics.Append(diags...)

	state := sysNetRoutes6Model{
		ID:     types.StringValue("net.routes6"),
		Routes: routesList,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
