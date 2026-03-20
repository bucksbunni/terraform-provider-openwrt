package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewSysNetConntrackDataSource() datasource.DataSource {
	return &sysNetConntrackDataSource{}
}

type sysNetConntrackDataSource struct {
	client *JsonRpcClient
}

type sysNetConntrackEntryModel struct {
	Proto      types.String `tfsdk:"proto"`
	SourceIP   types.String `tfsdk:"src_ip"`
	SourcePort types.Int64  `tfsdk:"src_port"`
	DestIP     types.String `tfsdk:"dest_ip"`
	DestPort   types.Int64  `tfsdk:"dest_port"`
	State      types.String `tfsdk:"state"`
}

type sysNetConntrackModel struct {
	ID      types.String `tfsdk:"id"`
	Entries types.List   `tfsdk:"entries"`
}

func (d *sysNetConntrackDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sys_net_conntrack"
}

func (d *sysNetConntrackDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves the connection tracking table from OpenWrt.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:    true,
				Description: "Internal ID.",
			},
			"entries": dsschema.ListAttribute{
				Computed:    true,
				Description: "List of tracked connections.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"proto":     types.StringType,
						"src_ip":    types.StringType,
						"src_port":  types.Int64Type,
						"dest_ip":   types.StringType,
						"dest_port": types.Int64Type,
						"state":     types.StringType,
					},
				},
			},
		},
	}
}

func (d *sysNetConntrackDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sysNetConntrackDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider must be configured before using this data source.")
		return
	}

	raw, err := d.client.SysCall(ctx, "net.conntrack")
	if err != nil {
		resp.Diagnostics.AddError("Error calling /rpc/sys", err.Error())
		return
	}

	var conntrackData []map[string]interface{}
	if err := json.Unmarshal(raw, &conntrackData); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	entries := make([]sysNetConntrackEntryModel, 0, len(conntrackData))
	for _, c := range conntrackData {
		cm := sysNetConntrackEntryModel{}
		if v, ok := c["proto"].(string); ok {
			cm.Proto = types.StringValue(v)
		}
		if v, ok := c["src"].(string); ok {
			cm.SourceIP = types.StringValue(v)
		}
		if v, ok := c["sport"].(float64); ok {
			cm.SourcePort = types.Int64Value(int64(v))
		}
		if v, ok := c["dest"].(string); ok {
			cm.DestIP = types.StringValue(v)
		}
		if v, ok := c["dport"].(float64); ok {
			cm.DestPort = types.Int64Value(int64(v))
		}
		if v, ok := c["state"].(string); ok {
			cm.State = types.StringValue(v)
		}
		entries = append(entries, cm)
	}

	entriesList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"proto":     types.StringType,
			"src_ip":    types.StringType,
			"src_port":  types.Int64Type,
			"dest_ip":   types.StringType,
			"dest_port": types.Int64Type,
			"state":     types.StringType,
		},
	}, entries)
	resp.Diagnostics.Append(diags...)

	state := sysNetConntrackModel{
		ID:      types.StringValue("net.conntrack"),
		Entries: entriesList,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
