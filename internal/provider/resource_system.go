package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewSystemResource() resource.Resource {
	return &systemResource{}
}

type systemResource struct {
	client *JsonRpcClient
}

type systemModel struct {
	ID           types.String `tfsdk:"id"`
	Hostname     types.String `tfsdk:"hostname"`
	TtyLogin     types.Bool   `tfsdk:"ttylogin"`
	LogSize      types.Int64  `tfsdk:"log_size"`
	UrandomSeed  types.Bool   `tfsdk:"urandom_seed"`
	Zonename     types.String `tfsdk:"zonename"`
	LogProto     types.String `tfsdk:"log_proto"`
	ConLogLevel  types.String `tfsdk:"conloglevel"`
	CronLogLevel types.String `tfsdk:"cronloglevel"`
}

func (r *systemResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system"
}

func (r *systemResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages OpenWrt system settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: system.",
			},
			"hostname": schema.StringAttribute{
				Optional:    true,
				Description: "System hostname.",
			},
			"ttylogin": schema.BoolAttribute{
				Optional:    true,
				Description: "Require TTY login.",
			},
			"log_size": schema.Int64Attribute{
				Optional:    true,
				Description: "System log size in KB.",
			},
			"urandom_seed": schema.BoolAttribute{
				Optional:    true,
				Description: "Seed urandom on boot.",
			},
			"zonename": schema.StringAttribute{
				Optional:    true,
				Description: "Timezone name (e.g., 'UTC').",
			},
			"log_proto": schema.StringAttribute{
				Optional:    true,
				Description: "Log protocol: 'udp', 'tcp'.",
			},
			"conloglevel": schema.StringAttribute{
				Optional:    true,
				Description: "Console log level.",
			},
			"cronloglevel": schema.StringAttribute{
				Optional:    true,
				Description: "Cron log level.",
			},
		},
	}
}

func (r *systemResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *systemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan systemModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := r.modelToOptions(plan)

	_, err := r.client.UCISection(ctx, "system", "system", "system", options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating system config", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "system"); err != nil {
		resp.Diagnostics.AddError("Error committing system config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue("system")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *systemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state systemModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data, err := r.client.UCIGetAll(ctx, "system", "system")
	if err != nil || len(data) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(data, &state)
	state.ID = types.StringValue("system")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *systemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan systemModel
	var state systemModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := r.modelToOptions(plan)

	if err := r.client.UCITSet(ctx, "system", "system", options); err != nil {
		resp.Diagnostics.AddError("Error updating system config", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "system"); err != nil {
		resp.Diagnostics.AddError("Error committing system config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *systemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Delete not supported", "System settings cannot be deleted, only modified.")
}

func (r *systemResource) modelToOptions(plan systemModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.Hostname.IsNull() {
		options["hostname"] = plan.Hostname.ValueString()
	}
	if !plan.TtyLogin.IsNull() {
		options["ttylogin"] = boolToString(plan.TtyLogin.ValueBool())
	}
	if !plan.LogSize.IsNull() {
		options["log_size"] = plan.LogSize.ValueInt64()
	}
	if !plan.UrandomSeed.IsNull() {
		options["urandom_seed"] = boolToString(plan.UrandomSeed.ValueBool())
	}
	if !plan.Zonename.IsNull() {
		options["zonename"] = plan.Zonename.ValueString()
	}
	if !plan.LogProto.IsNull() {
		options["log_proto"] = plan.LogProto.ValueString()
	}
	if !plan.ConLogLevel.IsNull() {
		options["conloglevel"] = plan.ConLogLevel.ValueString()
	}
	if !plan.CronLogLevel.IsNull() {
		options["cronloglevel"] = plan.CronLogLevel.ValueString()
	}

	return options
}

func (r *systemResource) optionsToModel(data map[string]interface{}, state *systemModel) {
	if v, ok := data["hostname"].(string); ok {
		state.Hostname = types.StringValue(v)
	}
	if v, ok := data["ttylogin"].(string); ok {
		state.TtyLogin = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["log_size"]; ok {
		if f, ok := v.(float64); ok {
			state.LogSize = types.Int64Value(int64(f))
		}
	}
	if v, ok := data["urandom_seed"].(string); ok {
		state.UrandomSeed = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["zonename"].(string); ok {
		state.Zonename = types.StringValue(v)
	}
	if v, ok := data["log_proto"].(string); ok {
		state.LogProto = types.StringValue(v)
	}
	if v, ok := data["conloglevel"].(string); ok {
		state.ConLogLevel = types.StringValue(v)
	}
	if v, ok := data["cronloglevel"].(string); ok {
		state.CronLogLevel = types.StringValue(v)
	}
}
