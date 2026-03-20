package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewSysInitDataSource() datasource.DataSource {
	return &sysInitDataSource{}
}

type sysInitDataSource struct {
	client *JsonRpcClient
}

type sysInitEntryModel struct {
	Name    types.String `tfsdk:"name"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Index   types.Int64  `tfsdk:"index"`
}

type sysInitModel struct {
	ID      types.String `tfsdk:"id"`
	Scripts types.List   `tfsdk:"scripts"`
}

func (d *sysInitDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sys_init"
}

func (d *sysInitDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves the list of init scripts from OpenWrt.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:    true,
				Description: "Internal ID.",
			},
			"scripts": dsschema.ListAttribute{
				Computed:    true,
				Description: "List of init scripts.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":    types.StringType,
						"enabled": types.BoolType,
						"index":   types.Int64Type,
					},
				},
			},
		},
	}
}

func (d *sysInitDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sysInitDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider must be configured before using this data source.")
		return
	}

	raw, err := d.client.SysCall(ctx, "init.names")
	if err != nil {
		resp.Diagnostics.AddError("Error calling /rpc/sys", err.Error())
		return
	}

	var namesData []string
	if err := json.Unmarshal(raw, &namesData); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	scripts := make([]sysInitEntryModel, 0, len(namesData))
	for _, name := range namesData {
		entry := sysInitEntryModel{
			Name: types.StringValue(name),
		}

		enabledRaw, err := d.client.SysCall(ctx, "init.enabled", name)
		if err == nil {
			var enabled bool
			if err := json.Unmarshal(enabledRaw, &enabled); err == nil {
				entry.Enabled = types.BoolValue(enabled)
			}
		}

		indexRaw, err := d.client.SysCall(ctx, "init.index", name)
		if err == nil {
			var index float64
			if err := json.Unmarshal(indexRaw, &index); err == nil {
				entry.Index = types.Int64Value(int64(index))
			}
		}

		scripts = append(scripts, entry)
	}

	scriptsList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":    types.StringType,
			"enabled": types.BoolType,
			"index":   types.Int64Type,
		},
	}, scripts)
	resp.Diagnostics.Append(diags...)

	state := sysInitModel{
		ID:      types.StringValue("init"),
		Scripts: scriptsList,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
