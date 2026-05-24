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

func NewNetworkDeviceResource() resource.Resource {
	return &networkDeviceResource{}
}

type networkDeviceResource struct {
	client *JsonRpcClient
}

type networkDeviceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	Ports          types.String `tfsdk:"ports"`
	Policy         types.String `tfsdk:"policy"`
	XmitHashPolicy types.String `tfsdk:"xmit_hash_policy"`
}

func (r *networkDeviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_device"
}

func (r *networkDeviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an OpenWrt network device (bridge, bonding, etc.).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: network/<device_name>.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Device name (e.g., 'br-lan', 'bond0').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Description: "Device type: 'bridge', 'bonding'.",
			},
			"ports": schema.StringAttribute{
				Optional:    true,
				Description: "Member ports, space-separated (e.g., 'eth0 eth1').",
			},
			"policy": schema.StringAttribute{
				Optional:    true,
				Description: "Bonding policy (e.g., '802.3ad', 'active-backup').",
			},
			"xmit_hash_policy": schema.StringAttribute{
				Optional:    true,
				Description: "Bonding transmit hash policy (e.g., 'layer2+3', 'layer3+4').",
			},
		},
	}
}

func (r *networkDeviceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *networkDeviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan networkDeviceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(plan)

	secName, err := r.client.UCIAdd(ctx, "network", "device")
	if err != nil {
		resp.Diagnostics.AddError("Error creating network device", err.Error())
		return
	}

	for key, value := range options {
		if err := r.client.UCISet(ctx, "network", secName, key, value); err != nil {
			resp.Diagnostics.AddError("Error setting network device option", err.Error())
			return
		}
	}

	if err := r.client.UCISet(ctx, "network", secName, "name", name); err != nil {
		resp.Diagnostics.AddError("Error setting network device name", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "network"); err != nil {
		resp.Diagnostics.AddError("Error committing network config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue(fmt.Sprintf("network/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *networkDeviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state networkDeviceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	devices, err := r.client.UCIForeach(ctx, "network", "device")
	if err != nil {
		resp.Diagnostics.AddError("Error reading network devices", err.Error())
		return
	}

	var data map[string]interface{}
	for _, dev := range devices {
		if dev["name"] == name {
			data = dev
			break
		}
	}

	if data == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(data, &state)
	state.ID = types.StringValue(fmt.Sprintf("network/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *networkDeviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan networkDeviceModel
	var state networkDeviceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(plan)

	devices, err := r.client.UCIForeach(ctx, "network", "device")
	if err != nil {
		resp.Diagnostics.AddError("Error reading network devices", err.Error())
		return
	}

	var secName string
	for _, dev := range devices {
		if dev["name"] == name {
			secName = dev[".name"].(string)
			break
		}
	}

	if secName == "" {
		resp.Diagnostics.AddError("Error finding network device", "device not found")
		return
	}

	for key, value := range options {
		if err := r.client.UCISet(ctx, "network", secName, key, value); err != nil {
			resp.Diagnostics.AddError("Error updating network device option", err.Error())
			return
		}
	}

	if err := r.client.UCICommit(ctx, "network"); err != nil {
		resp.Diagnostics.AddError("Error committing network config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue(fmt.Sprintf("network/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *networkDeviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state networkDeviceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	devices, err := r.client.UCIForeach(ctx, "network", "device")
	if err != nil {
		resp.Diagnostics.AddError("Error reading network devices", err.Error())
		return
	}

	var secName string
	for _, dev := range devices {
		if dev["name"] == name {
			secName = dev[".name"].(string)
			break
		}
	}

	if secName != "" {
		if err := r.client.UCIDelete(ctx, "network", secName); err != nil {
			resp.Diagnostics.AddError("Error deleting network device", err.Error())
			return
		}
		if err := r.client.UCICommit(ctx, "network"); err != nil {
			resp.Diagnostics.AddError("Error committing network config", err.Error())
			return
		}
		if err := r.client.UCIApply(ctx, false); err != nil {
			tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
		}
	}

	resp.State.RemoveResource(ctx)
}

func (r *networkDeviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] != "network" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: network/<device_name>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}

func (r *networkDeviceResource) modelToOptions(plan networkDeviceModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.Type.IsNull() {
		options["type"] = plan.Type.ValueString()
	}
	if !plan.Ports.IsNull() {
		options["ports"] = plan.Ports.ValueString()
	}
	if !plan.Policy.IsNull() {
		options["policy"] = plan.Policy.ValueString()
	}
	if !plan.XmitHashPolicy.IsNull() {
		options["xmit_hash_policy"] = plan.XmitHashPolicy.ValueString()
	}

	return options
}

func (r *networkDeviceResource) optionsToModel(data map[string]interface{}, state *networkDeviceModel) {
	if v, ok := data["type"].(string); ok {
		state.Type = types.StringValue(v)
	}
	if v, ok := data["name"].(string); ok {
		state.Name = types.StringValue(v)
	}
	if v, ok := data["ports"].(string); ok {
		state.Ports = types.StringValue(v)
	}
	if v, ok := data["policy"].(string); ok {
		state.Policy = types.StringValue(v)
	}
	if v, ok := data["xmit_hash_policy"].(string); ok {
		state.XmitHashPolicy = types.StringValue(v)
	}
}
