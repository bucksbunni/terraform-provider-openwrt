package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewFirewallDefaultsResource() resource.Resource {
	return &firewallDefaultsResource{}
}

type firewallDefaultsResource struct {
	client *JsonRpcClient
}

type firewallDefaultsModel struct {
	ID              types.String `tfsdk:"id"`
	Input           types.String `tfsdk:"input"`
	Output          types.String `tfsdk:"output"`
	Forward         types.String `tfsdk:"forward"`
	SynfloodProtect types.Bool   `tfsdk:"synflood_protect"`
	SynfloodRate    types.String `tfsdk:"synflood_rate"`
	DropInvalid     types.Bool   `tfsdk:"drop_invalid"`
	AutoHelper      types.Bool   `tfsdk:"auto_helper"`
}

func (r *firewallDefaultsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_defaults"
}

func (r *firewallDefaultsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages OpenWrt firewall default policies.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: firewall/defaults.",
			},
			"input": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Input policy: 'ACCEPT', 'REJECT', 'DROP'.",
			},
			"output": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Output policy: 'ACCEPT', 'REJECT', 'DROP'.",
			},
			"forward": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Forward policy: 'ACCEPT', 'REJECT', 'DROP'.",
			},
			"synflood_protect": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable SYN flood protection.",
			},
			"synflood_rate": schema.StringAttribute{
				Optional:    true,
				Description: "SYN flood rate limit (e.g., '25/s').",
			},
			"drop_invalid": schema.BoolAttribute{
				Optional:    true,
				Description: "Drop invalid packets.",
			},
			"auto_helper": schema.BoolAttribute{
				Optional:    true,
				Description: "Automatically create helper rules for active connections.",
			},
		},
	}
}

func (r *firewallDefaultsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *firewallDefaultsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan firewallDefaultsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := r.modelToOptions(plan)

	_, err := r.client.UCISection(ctx, "firewall", "defaults", "defaults", options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating firewall defaults", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "firewall"); err != nil {
		resp.Diagnostics.AddError("Error committing firewall config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue("firewall/defaults")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *firewallDefaultsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state firewallDefaultsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data, err := r.client.UCIGetAll(ctx, "firewall", "defaults")
	if err != nil || len(data) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(data, &state)
	state.ID = types.StringValue("firewall/defaults")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallDefaultsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan firewallDefaultsModel
	var state firewallDefaultsModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := r.modelToOptions(plan)

	if err := r.client.UCITSet(ctx, "firewall", "defaults", options); err != nil {
		resp.Diagnostics.AddError("Error updating firewall defaults", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "firewall"); err != nil {
		resp.Diagnostics.AddError("Error committing firewall config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *firewallDefaultsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Delete not supported", "Firewall defaults cannot be deleted, only modified.")
}

func (r *firewallDefaultsResource) modelToOptions(plan firewallDefaultsModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.Input.IsNull() {
		options["input"] = plan.Input.ValueString()
	}
	if !plan.Output.IsNull() {
		options["output"] = plan.Output.ValueString()
	}
	if !plan.Forward.IsNull() {
		options["forward"] = plan.Forward.ValueString()
	}
	if !plan.SynfloodProtect.IsNull() {
		options["synflood_protect"] = boolToString(plan.SynfloodProtect.ValueBool())
	}
	if !plan.SynfloodRate.IsNull() {
		options["synflood_rate"] = plan.SynfloodRate.ValueString()
	}
	if !plan.DropInvalid.IsNull() {
		options["drop_invalid"] = boolToString(plan.DropInvalid.ValueBool())
	}
	if !plan.AutoHelper.IsNull() {
		options["auto_helper"] = boolToString(plan.AutoHelper.ValueBool())
	}

	return options
}

func (r *firewallDefaultsResource) optionsToModel(data map[string]interface{}, state *firewallDefaultsModel) {
	if v, ok := data["input"].(string); ok {
		state.Input = types.StringValue(v)
	}
	if v, ok := data["output"].(string); ok {
		state.Output = types.StringValue(v)
	}
	if v, ok := data["forward"].(string); ok {
		state.Forward = types.StringValue(v)
	}
	if v, ok := data["synflood_protect"].(string); ok {
		state.SynfloodProtect = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["synflood_rate"].(string); ok {
		state.SynfloodRate = types.StringValue(v)
	}
	if v, ok := data["drop_invalid"].(string); ok {
		state.DropInvalid = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["auto_helper"].(string); ok {
		state.AutoHelper = types.BoolValue(v == "1" || v == "true")
	}
}
