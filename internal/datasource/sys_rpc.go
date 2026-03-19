package datasource

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bucksbunni/terraform-provider-openwrt/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &sysRPCDataSource{}

func NewSysRPCDataSource() datasource.DataSource {
	return &sysRPCDataSource{}
}

type sysRPCDataSource struct {
	client *client.JsonRpcClient
}

type sysRPCModel struct {
	ID         types.String `tfsdk:"id"`
	Method     types.String `tfsdk:"method"`
	ParamsJSON types.String `tfsdk:"params_json"`
	ResultJSON types.String `tfsdk:"result_json"`
}

func (d *sysRPCDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sys_rpc"
}

func (d *sysRPCDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Low-level access to the /rpc/sys JSON-RPC API. " +
			"Allows calling any luci.sys* function; e.g. 'hostname', 'uptime', " +
			"'net.routes', 'user.getuser', 'process.list', 'wifi.getiwinfo', etc.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:    true,
				Description: "Internal ID, equal to the method name.",
			},
			"method": dsschema.StringAttribute{
				Required: true,
				Description: "JSON-RPC method name to call on /rpc/sys, " +
					"e.g. 'hostname', 'uptime', 'net.routes', 'user.getuser'.",
			},
			"params_json": dsschema.StringAttribute{
				Optional: true,
				Description: "Optional JSON-encoded array of positional parameters " +
					"to pass to the method, e.g. '[\"network\",\"lan\"]'.",
			},
			"result_json": dsschema.StringAttribute{
				Computed: true,
				Description: "Raw JSON result returned by the RPC call. " +
					"Use jsondecode() in Terraform to work with it.",
			},
		},
	}
}

func (d *sysRPCDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.JsonRpcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.JsonRpcClient, got %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *sysRPCDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider must be configured before using this data source.",
		)
		return
	}

	var config sysRPCModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	method := config.Method.ValueString()
	if method == "" {
		resp.Diagnostics.AddError("Missing method", "`method` must be set.")
		return
	}

	var params []interface{}
	if !config.ParamsJSON.IsNull() && config.ParamsJSON.ValueString() != "" {
		if err := json.Unmarshal([]byte(config.ParamsJSON.ValueString()), &params); err != nil {
			resp.Diagnostics.AddError(
				"Invalid params_json",
				fmt.Sprintf("params_json must be a JSON array: %v", err),
			)
			return
		}
	}

	raw, err := d.client.SysCall(ctx, method, params...)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error calling /rpc/sys",
			err.Error(),
		)
		return
	}

	state := sysRPCModel{
		ID:         types.StringValue(method),
		Method:     types.StringValue(method),
		ParamsJSON: config.ParamsJSON,
		ResultJSON: types.StringValue(string(raw)),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
