package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewSysProcessListDataSource() datasource.DataSource {
	return &sysProcessListDataSource{}
}

type sysProcessListDataSource struct {
	client *JsonRpcClient
}

type sysProcessModel struct {
	PID     types.Int64  `tfsdk:"pid"`
	PPID    types.Int64  `tfsdk:"ppid"`
	PGID    types.Int64  `tfsdk:"pgid"`
	Name    types.String `tfsdk:"name"`
	URandom types.String `tfsdk:"urandom"`
	State   types.String `tfsdk:"state"`
	User    types.String `tfsdk:"user"`
}

type sysProcessListModel struct {
	ID        types.String `tfsdk:"id"`
	Processes types.List   `tfsdk:"processes"`
}

func (d *sysProcessListDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sys_process_list"
}

func (d *sysProcessListDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves the list of running processes from OpenWrt.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:    true,
				Description: "Internal ID.",
			},
			"processes": dsschema.ListAttribute{
				Computed:    true,
				Description: "List of running processes.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"pid":     types.Int64Type,
						"ppid":    types.Int64Type,
						"pgid":    types.Int64Type,
						"name":    types.StringType,
						"urandom": types.StringType,
						"state":   types.StringType,
						"user":    types.StringType,
					},
				},
			},
		},
	}
}

func (d *sysProcessListDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sysProcessListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider must be configured before using this data source.")
		return
	}

	raw, err := d.client.SysCall(ctx, "process.list")
	if err != nil {
		resp.Diagnostics.AddError("Error calling /rpc/sys", err.Error())
		return
	}

	var processData []map[string]interface{}
	if err := json.Unmarshal(raw, &processData); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	processes := make([]sysProcessModel, 0, len(processData))
	for _, p := range processData {
		pm := sysProcessModel{}
		if v, ok := p["PID"].(float64); ok {
			pm.PID = types.Int64Value(int64(v))
		}
		if v, ok := p["PPID"].(float64); ok {
			pm.PPID = types.Int64Value(int64(v))
		}
		if v, ok := p["PGID"].(float64); ok {
			pm.PGID = types.Int64Value(int64(v))
		}
		if v, ok := p["NAME"].(string); ok {
			pm.Name = types.StringValue(v)
		}
		if v, ok := p["URANDOM"].(string); ok {
			pm.URandom = types.StringValue(v)
		}
		if v, ok := p["STAT"].(string); ok {
			pm.State = types.StringValue(v)
		}
		if v, ok := p["USER"].(string); ok {
			pm.User = types.StringValue(v)
		}
		processes = append(processes, pm)
	}

	processesList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"pid":     types.Int64Type,
			"ppid":    types.Int64Type,
			"pgid":    types.Int64Type,
			"name":    types.StringType,
			"urandom": types.StringType,
			"state":   types.StringType,
			"user":    types.StringType,
		},
	}, processes)
	resp.Diagnostics.Append(diags...)

	state := sysProcessListModel{
		ID:        types.StringValue("process.list"),
		Processes: processesList,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
