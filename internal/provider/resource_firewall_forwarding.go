package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewFirewallForwardingResource() resource.Resource {
	return &firewallForwardingResource{}
}

type firewallForwardingResource struct {
	client *JsonRpcClient
}

type firewallForwardingModel struct {
	ID   types.String `tfsdk:"id"`
	Src  types.String `tfsdk:"src"`
	Dest types.String `tfsdk:"dest"`
}

func (r *firewallForwardingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_forwarding"
}

func (r *firewallForwardingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an OpenWrt firewall forwarding rule (zone-to-zone traffic).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: firewall/<src>_<dest>.",
			},
			"src": schema.StringAttribute{
				Required:    true,
				Description: "Source zone name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dest": schema.StringAttribute{
				Required:    true,
				Description: "Destination zone name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *firewallForwardingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *firewallForwardingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan firewallForwardingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	src := plan.Src.ValueString()
	dest := plan.Dest.ValueString()

	forwardName := fmt.Sprintf("%s_%s", src, dest)

	options := map[string]interface{}{
		"src":  src,
		"dest": dest,
	}

	_, err := r.client.UCISection(ctx, "firewall", "forwarding", forwardName, options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating firewall forwarding", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "firewall"); err != nil {
		resp.Diagnostics.AddError("Error committing firewall config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue(fmt.Sprintf("firewall/%s", forwardName))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *firewallForwardingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state firewallForwardingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	src := state.Src.ValueString()
	dest := state.Dest.ValueString()
	forwardName := fmt.Sprintf("%s_%s", src, dest)

	data, err := r.client.UCIGetAll(ctx, "firewall", forwardName)
	if err != nil || len(data) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(data, &state)
	state.ID = types.StringValue(fmt.Sprintf("firewall/%s", forwardName))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallForwardingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan firewallForwardingModel
	var state firewallForwardingModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	src := plan.Src.ValueString()
	dest := plan.Dest.ValueString()
	forwardName := fmt.Sprintf("%s_%s", src, dest)

	options := map[string]interface{}{
		"src":  src,
		"dest": dest,
	}

	if err := r.client.UCITSet(ctx, "firewall", forwardName, options); err != nil {
		resp.Diagnostics.AddError("Error updating firewall forwarding", err.Error())
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

func (r *firewallForwardingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state firewallForwardingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	src := state.Src.ValueString()
	dest := state.Dest.ValueString()
	forwardName := fmt.Sprintf("%s_%s", src, dest)

	if err := r.client.UCIDelete(ctx, "firewall", forwardName); err != nil {
		resp.Diagnostics.AddError("Error deleting firewall forwarding", err.Error())
		return
	}
	if err := r.client.UCICommit(ctx, "firewall"); err != nil {
		resp.Diagnostics.AddError("Error committing firewall config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.State.RemoveResource(ctx)
}

func (r *firewallForwardingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] != "firewall" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: firewall/<src>_<dest>")
		return
	}

	forwardParts := strings.SplitN(parts[1], "_", 2)
	if len(forwardParts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: firewall/<src>_<dest>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("src"), forwardParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("dest"), forwardParts[1])...)
}

func (r *firewallForwardingResource) optionsToModel(data map[string]interface{}, state *firewallForwardingModel) {
	if v, ok := data["src"].(string); ok {
		state.Src = types.StringValue(v)
	}
	if v, ok := data["dest"].(string); ok {
		state.Dest = types.StringValue(v)
	}
}
