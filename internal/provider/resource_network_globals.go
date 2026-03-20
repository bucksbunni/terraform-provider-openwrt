package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewNetworkGlobalsResource() resource.Resource {
	return &networkGlobalsResource{}
}

type networkGlobalsResource struct {
	client *JsonRpcClient
}

type networkGlobalsModel struct {
	ID             types.String `tfsdk:"id"`
	ULAPrefix      types.String `tfsdk:"ula_prefix"`
	PacketSteering types.Bool   `tfsdk:"packet_steering"`
}

func (r *networkGlobalsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_globals"
}

func (r *networkGlobalsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages OpenWrt network global settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: network/globals.",
			},
			"ula_prefix": schema.StringAttribute{
				Optional:    true,
				Description: "IPv6 ULA prefix (e.g., 'fda1:b7f5:46f7::/48').",
			},
			"packet_steering": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable packet steering across CPU cores.",
			},
		},
	}
}

func (r *networkGlobalsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *networkGlobalsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan networkGlobalsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := r.modelToOptions(plan)

	_, err := r.client.UCISection(ctx, "network", "globals", "globals", options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating network globals", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "network"); err != nil {
		resp.Diagnostics.AddError("Error committing network config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue("network/globals")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *networkGlobalsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state networkGlobalsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data, err := r.client.UCIGetAll(ctx, "network", "globals")
	if err != nil || len(data) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(data, &state)
	state.ID = types.StringValue("network/globals")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *networkGlobalsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan networkGlobalsModel
	var state networkGlobalsModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := r.modelToOptions(plan)

	if err := r.client.UCITSet(ctx, "network", "globals", options); err != nil {
		resp.Diagnostics.AddError("Error updating network globals", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "network"); err != nil {
		resp.Diagnostics.AddError("Error committing network config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *networkGlobalsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Delete not supported", "Network globals cannot be deleted, only modified.")
}

func (r *networkGlobalsResource) modelToOptions(plan networkGlobalsModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.ULAPrefix.IsNull() {
		options["ula_prefix"] = plan.ULAPrefix.ValueString()
	}
	if !plan.PacketSteering.IsNull() {
		options["packet_steering"] = boolToString(plan.PacketSteering.ValueBool())
	}

	return options
}

func (r *networkGlobalsResource) optionsToModel(data map[string]interface{}, state *networkGlobalsModel) {
	if v, ok := data["ula_prefix"].(string); ok {
		state.ULAPrefix = types.StringValue(v)
	}
	if v, ok := data["packet_steering"].(string); ok {
		state.PacketSteering = types.BoolValue(v == "1" || v == "true")
	}
}
