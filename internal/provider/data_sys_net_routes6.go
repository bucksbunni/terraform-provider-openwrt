package provider

import (
	"context"
	"encoding/json"

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

	raw, err := d.client.SysCall(ctx, "net.routes6")
	if err != nil {
		resp.Diagnostics.AddError("Error calling /rpc/sys", err.Error())
		return
	}

	var routesData []map[string]interface{}
	if err := json.Unmarshal(raw, &routesData); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	routes := make([]sysNetRoute6Model, 0, len(routesData))
	for _, r := range routesData {
		rm := sysNetRoute6Model{}
		if v, ok := r["source"].(string); ok {
			rm.Source = types.StringValue(v)
		}
		if v, ok := r["dest"].(string); ok {
			rm.Dest = types.StringValue(v)
		}
		if v, ok := r["nexthop"].(string); ok {
			rm.Nexthop = types.StringValue(v)
		}
		if v, ok := r["metric"].(float64); ok {
			rm.Metric = types.Int64Value(int64(v))
		}
		if v, ok := r["refcount"].(float64); ok {
			rm.RefCount = types.Int64Value(int64(v))
		}
		if v, ok := r["usecount"].(float64); ok {
			rm.UseCount = types.Int64Value(int64(v))
		}
		if v, ok := r["flags"].(string); ok {
			rm.Flags = types.StringValue(v)
		}
		if v, ok := r["device"].(string); ok {
			rm.Device = types.StringValue(v)
		}
		routes = append(routes, rm)
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
