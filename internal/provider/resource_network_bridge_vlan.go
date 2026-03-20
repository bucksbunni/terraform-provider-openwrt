package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewNetworkBridgeVlanResource() resource.Resource {
	return &networkBridgeVlanResource{}
}

type networkBridgeVlanResource struct {
	client *JsonRpcClient
}

type networkBridgeVlanModel struct {
	ID     types.String `tfsdk:"id"`
	Device types.String `tfsdk:"device"`
	VLAN   types.Int64  `tfsdk:"vlan"`
	Ports  types.String `tfsdk:"ports"`
}

func (r *networkBridgeVlanResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_bridge_vlan"
}

func (r *networkBridgeVlanResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a bridge VLAN assignment in OpenWrt.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: network/<device>_<vlan>.",
			},
			"device": schema.StringAttribute{
				Required:    true,
				Description: "Bridge device name (e.g., 'br-lan').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vlan": schema.Int64Attribute{
				Required:    true,
				Description: "VLAN ID (1-4094).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"ports": schema.StringAttribute{
				Required:    true,
				Description: "Ports for this VLAN, space-separated (e.g., 'eth0:u* eth1:t').",
			},
		},
	}
}

func (r *networkBridgeVlanResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *networkBridgeVlanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan networkBridgeVlanModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	device := plan.Device.ValueString()
	vlan := plan.VLAN.ValueInt64()
	vlanName := fmt.Sprintf("%s_%d", device, vlan)

	options := r.modelToOptions(plan)

	_, err := r.client.UCISection(ctx, "network", "bridge-vlan", vlanName, options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating bridge VLAN", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "network"); err != nil {
		resp.Diagnostics.AddError("Error committing network config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue(fmt.Sprintf("network/%s", vlanName))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *networkBridgeVlanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state networkBridgeVlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	device := state.Device.ValueString()
	vlan := state.VLAN.ValueInt64()
	vlanName := fmt.Sprintf("%s_%d", device, vlan)

	data, err := r.client.UCIGetAll(ctx, "network", vlanName)
	if err != nil || len(data) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(data, &state)
	state.ID = types.StringValue(fmt.Sprintf("network/%s", vlanName))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *networkBridgeVlanResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan networkBridgeVlanModel
	var state networkBridgeVlanModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	device := plan.Device.ValueString()
	vlan := plan.VLAN.ValueInt64()
	vlanName := fmt.Sprintf("%s_%d", device, vlan)

	options := r.modelToOptions(plan)

	if err := r.client.UCITSet(ctx, "network", vlanName, options); err != nil {
		resp.Diagnostics.AddError("Error updating bridge VLAN", err.Error())
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

func (r *networkBridgeVlanResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state networkBridgeVlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	device := state.Device.ValueString()
	vlan := state.VLAN.ValueInt64()
	vlanName := fmt.Sprintf("%s_%d", device, vlan)

	if err := r.client.UCIDelete(ctx, "network", vlanName); err != nil {
		resp.Diagnostics.AddError("Error deleting bridge VLAN", err.Error())
		return
	}
	if err := r.client.UCICommit(ctx, "network"); err != nil {
		resp.Diagnostics.AddError("Error committing network config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.State.RemoveResource(ctx)
}

func (r *networkBridgeVlanResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] != "network" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: network/<device>_<vlan>")
		return
	}

	vlanParts := strings.SplitN(parts[1], "_", 2)
	if len(vlanParts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: network/<device>_<vlan>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device"), vlanParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vlan"), vlanParts[1])...)
}

func (r *networkBridgeVlanResource) modelToOptions(plan networkBridgeVlanModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.Device.IsNull() {
		options["device"] = plan.Device.ValueString()
	}
	if !plan.VLAN.IsNull() {
		options["vlan"] = plan.VLAN.ValueInt64()
	}
	if !plan.Ports.IsNull() {
		options["ports"] = plan.Ports.ValueString()
	}

	return options
}

func (r *networkBridgeVlanResource) optionsToModel(data map[string]interface{}, state *networkBridgeVlanModel) {
	if v, ok := data["device"].(string); ok {
		state.Device = types.StringValue(v)
	}
	if v, ok := data["vlan"]; ok {
		if f, ok := v.(float64); ok {
			state.VLAN = types.Int64Value(int64(f))
		}
	}
	if v, ok := data["ports"].(string); ok {
		state.Ports = types.StringValue(v)
	}
}
