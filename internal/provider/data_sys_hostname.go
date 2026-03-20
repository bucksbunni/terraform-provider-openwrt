package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewSysHostnameDataSource() datasource.DataSource {
	return &sysHostnameDataSource{}
}

type sysHostnameDataSource struct {
	client *JsonRpcClient
}

type sysHostnameModel struct {
	ID       types.String `tfsdk:"id"`
	Hostname types.String `tfsdk:"hostname"`
}

func (d *sysHostnameDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sys_hostname"
}

func (d *sysHostnameDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves the system hostname from OpenWrt.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:    true,
				Description: "Internal ID.",
			},
			"hostname": dsschema.StringAttribute{
				Computed:    true,
				Description: "Current system hostname.",
			},
		},
	}
}

func (d *sysHostnameDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sysHostnameDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider must be configured before using this data source.")
		return
	}

	raw, err := d.client.SysCall(ctx, "hostname")
	if err != nil {
		resp.Diagnostics.AddError("Error calling /rpc/sys", err.Error())
		return
	}

	var hostname string
	if err := json.Unmarshal(raw, &hostname); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	state := sysHostnameModel{
		ID:       types.StringValue("hostname"),
		Hostname: types.StringValue(hostname),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
