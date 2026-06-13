package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewSystemNTPResource() resource.Resource {
	return &systemNTPResource{}
}

type systemNTPResource struct {
	client *JsonRpcClient
}

type systemNTPModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Server  types.List   `tfsdk:"server"`
	Port    types.Int64  `tfsdk:"port"`
}

func (r *systemNTPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_ntp"
}

func (r *systemNTPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages OpenWrt NTP (timeserver) configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: system/ntp.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Instance name (e.g., 'ntp').",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable NTP client.",
			},
			"server": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "NTP servers (e.g., ['0.openwrt.pool.ntp.org', '1.openwrt.pool.ntp.org']).",
			},
			"port": schema.Int64Attribute{
				Optional:    true,
				Description: "NTP port (default: 123).",
			},
		},
	}
}

func (r *systemNTPResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*JsonRpcClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *JsonRpcClient, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *systemNTPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan systemNTPModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(ctx, plan)

	_, err := r.client.UCISection(ctx, "system", "timeserver", name, options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating NTP config", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "system"); err != nil {
		resp.Diagnostics.AddError("Error committing system config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue(fmt.Sprintf("system/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *systemNTPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state systemNTPModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	data, err := r.client.UCIGetAll(ctx, "system", name)
	if err != nil || len(data) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(data, &state)
	state.ID = types.StringValue(fmt.Sprintf("system/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *systemNTPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan systemNTPModel
	var state systemNTPModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(ctx, plan)

	if err := r.client.UCITSet(ctx, "system", name, options); err != nil {
		resp.Diagnostics.AddError("Error updating NTP config", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "system"); err != nil {
		resp.Diagnostics.AddError("Error committing system config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue(fmt.Sprintf("system/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *systemNTPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] != "system" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: system/<name>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}

func (r *systemNTPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state systemNTPModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	if err := r.client.UCIDelete(ctx, "system", name); err != nil {
		resp.Diagnostics.AddError("Error deleting NTP config", err.Error())
		return
	}
	if err := r.client.UCICommit(ctx, "system"); err != nil {
		resp.Diagnostics.AddError("Error committing system config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.State.RemoveResource(ctx)
}

func (r *systemNTPResource) modelToOptions(ctx context.Context, plan systemNTPModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.Enabled.IsNull() {
		options["enabled"] = boolToString(plan.Enabled.ValueBool())
	}
	if !plan.Server.IsNull() {
		var serverList []string
		diagnostics := plan.Server.ElementsAs(ctx, &serverList, false)
		if !diagnostics.HasError() {
			options["server"] = strings.Join(serverList, " ")
		}
	}
	if !plan.Port.IsNull() {
		options["port"] = plan.Port.ValueInt64()
	}

	return options
}

func (r *systemNTPResource) optionsToModel(data map[string]interface{}, state *systemNTPModel) {
	if v, ok := data["enabled"].(string); ok {
		state.Enabled = types.BoolValue(v == "1" || v == "true")
	}
	if server, ok := data["server"]; ok {
		switch v := server.(type) {
		case []interface{}:
			var serverList []string
			for _, s := range v {
				if str, ok := s.(string); ok {
					serverList = append(serverList, str)
				}
			}
			state.Server, _ = types.ListValueFrom(context.Background(), types.StringType, serverList)
		case string:
			if v != "" {
				serverList := strings.Split(v, " ")
				state.Server, _ = types.ListValueFrom(context.Background(), types.StringType, serverList)
			}
		}
	}
	if v, ok := data["port"]; ok {
		if f, ok := v.(float64); ok {
			state.Port = types.Int64Value(int64(f))
		}
	}
}
