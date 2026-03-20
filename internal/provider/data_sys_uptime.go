package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewSysUptimeDataSource() datasource.DataSource {
	return &sysUptimeDataSource{}
}

type sysUptimeDataSource struct {
	client *JsonRpcClient
}

type sysUptimeModel struct {
	ID     types.String `tfsdk:"id"`
	Uptime types.String `tfsdk:"uptime"`
}

func (d *sysUptimeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sys_uptime"
}

func (d *sysUptimeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves the system uptime from OpenWrt.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:    true,
				Description: "Internal ID.",
			},
			"uptime": dsschema.StringAttribute{
				Computed:    true,
				Description: "System uptime in seconds.",
			},
		},
	}
}

func (d *sysUptimeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sysUptimeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider must be configured before using this data source.")
		return
	}

	raw, err := d.client.SysCall(ctx, "uptime")
	if err != nil {
		resp.Diagnostics.AddError("Error calling /rpc/sys", err.Error())
		return
	}

	var uptime string
	if err := json.Unmarshal(raw, &uptime); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	state := sysUptimeModel{
		ID:     types.StringValue("uptime"),
		Uptime: types.StringValue(uptime),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
