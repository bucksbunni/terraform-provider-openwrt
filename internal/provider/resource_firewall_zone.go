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

func NewFirewallZoneResource() resource.Resource {
	return &firewallZoneResource{}
}

type firewallZoneResource struct {
	client *JsonRpcClient
}

type firewallZoneModel struct {
	ID          types.String `tfsdk:"id"`
	SectionName types.String `tfsdk:"section_name"`
	Name        types.String `tfsdk:"name"`
	Input       types.String `tfsdk:"input"`
	Output      types.String `tfsdk:"output"`
	Forward     types.String `tfsdk:"forward"`
	Masq        types.Bool   `tfsdk:"masq"`
	MasqSrc     types.String `tfsdk:"masq_src"`
	MasqDest    types.String `tfsdk:"masq_dest"`
	MtuFix      types.Bool   `tfsdk:"mtu_fix"`
	Network     types.List   `tfsdk:"network"`
}

func (r *firewallZoneResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_zone"
}

func (r *firewallZoneResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an OpenWrt firewall zone.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: firewall/<zone_name>.",
			},
			"section_name": schema.StringAttribute{
				Computed:    true,
				Description: "Internal UCI section identifier (e.g. 'cfg0abc12') of the anonymous section backing this resource. Managed automatically; resolved on import.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the firewall zone (e.g., 'lan', 'wan', 'guest').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
			"masq": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable masquerading (NAT) for this zone.",
			},
			"masq_src": schema.StringAttribute{
				Optional:    true,
				Description: "Source addresses to masquerade (CIDR).",
			},
			"masq_dest": schema.StringAttribute{
				Optional:    true,
				Description: "Destination addresses to masquerade (CIDR).",
			},
			"mtu_fix": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable MSS clamping for outgoing traffic.",
			},
			"network": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Network interface names belonging to this zone (e.g., ['lan', 'guest']).",
			},
		},
	}
}

func (r *firewallZoneResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *firewallZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan firewallZoneModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(ctx, plan)

	tflog.Debug(ctx, "Creating firewall zone", map[string]interface{}{"name": name, "options": options})

	// For firewall zones, we need to use add to create an anonymous section first
	// then set the name option
	secName, err := r.client.UCIAdd(ctx, "firewall", "zone")
	if err != nil {
		resp.Diagnostics.AddError("Error creating firewall zone", err.Error())
		return
	}
	plan.SectionName = types.StringValue(secName)

	tflog.Debug(ctx, "UCIAdd returned", map[string]interface{}{"secName": secName})

	// Now set all the options including name
	for key, value := range options {
		tflog.Debug(ctx, "UCISet", map[string]interface{}{"key": key, "value": value, "secName": secName})
		if err := r.client.UCISet(ctx, "firewall", secName, key, value); err != nil {
			resp.Diagnostics.AddError("Error setting firewall zone option", err.Error())
			return
		}
	}

	// Set the name explicitly
	if err := r.client.UCISet(ctx, "firewall", secName, "name", name); err != nil {
		resp.Diagnostics.AddError("Error setting firewall zone name", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "firewall"); err != nil {
		resp.Diagnostics.AddError("Error committing firewall config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue(fmt.Sprintf("firewall/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *firewallZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state firewallZoneModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	tflog.Debug(ctx, "Read firewall zone", map[string]interface{}{"name": name, "id": state.ID.ValueString()})

	data, sectionName, found, err := r.client.UCIResolveSection(ctx, "firewall", "zone", state.SectionName.ValueString(), map[string]string{"name": name})
	if err != nil {
		resp.Diagnostics.AddError("Error reading firewall zones", err.Error())
		return
	}
	if !found {
		tflog.Debug(ctx, "Firewall zone not found, removing", map[string]interface{}{"name": name})
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Debug(ctx, "Firewall zone data", map[string]interface{}{"data": data})

	r.optionsToModel(ctx, data, &state)
	state.SectionName = types.StringValue(sectionName)
	state.ID = types.StringValue(fmt.Sprintf("firewall/%s", name))

	tflog.Debug(ctx, "Firewall zone state after read", map[string]interface{}{"name": state.Name.ValueString(), "input": state.Input.ValueString(), "output": state.Output.ValueString(), "forward": state.Forward.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan firewallZoneModel
	var state firewallZoneModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(ctx, plan)

	_, secName, found, err := r.client.UCIResolveSection(ctx, "firewall", "zone", state.SectionName.ValueString(), map[string]string{"name": name})
	if err != nil {
		resp.Diagnostics.AddError("Error reading firewall zones", err.Error())
		return
	}
	if !found {
		resp.Diagnostics.AddError("Error finding firewall zone", "zone not found")
		return
	}

	for key, value := range options {
		if err := r.client.UCISet(ctx, "firewall", secName, key, value); err != nil {
			resp.Diagnostics.AddError("Error updating firewall zone option", err.Error())
			return
		}
	}

	if err := r.client.UCICommit(ctx, "firewall"); err != nil {
		resp.Diagnostics.AddError("Error committing firewall config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.SectionName = types.StringValue(secName)
	plan.ID = types.StringValue(fmt.Sprintf("firewall/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *firewallZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state firewallZoneModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	_, secName, found, err := r.client.UCIResolveSection(ctx, "firewall", "zone", state.SectionName.ValueString(), map[string]string{"name": name})
	if err != nil {
		resp.Diagnostics.AddError("Error reading firewall zones", err.Error())
		return
	}
	if !found {
		// Section already absent - treat as already deleted.
		resp.State.RemoveResource(ctx)
		return
	}

	if err := r.client.UCIDelete(ctx, "firewall", secName); err != nil {
		resp.Diagnostics.AddError("Error deleting firewall zone", err.Error())
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

func (r *firewallZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] != "firewall" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: firewall/<zone_name>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}

func (r *firewallZoneResource) modelToOptions(ctx context.Context, plan firewallZoneModel) map[string]interface{} {
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
	if !plan.Masq.IsNull() {
		if plan.Masq.ValueBool() {
			options["masq"] = "1"
		} else {
			options["masq"] = "0"
		}
	}
	if !plan.MasqSrc.IsNull() {
		options["masq_src"] = plan.MasqSrc.ValueString()
	}
	if !plan.MasqDest.IsNull() {
		options["masq_dest"] = plan.MasqDest.ValueString()
	}
	if !plan.MtuFix.IsNull() {
		if plan.MtuFix.ValueBool() {
			options["mtu_fix"] = "1"
		} else {
			options["mtu_fix"] = "0"
		}
	}
	if !plan.Network.IsNull() {
		var networks []string
		plan.Network.ElementsAs(ctx, &networks, false)
		options["network"] = strings.Join(networks, " ")
	}

	return options
}

func (r *firewallZoneResource) optionsToModel(ctx context.Context, data map[string]interface{}, state *firewallZoneModel) {
	if v, ok := data["name"].(string); ok {
		state.Name = types.StringValue(v)
	}
	if v, ok := data["input"].(string); ok {
		state.Input = types.StringValue(v)
	} else {
		state.Input = types.StringValue("ACCEPT")
	}
	if v, ok := data["output"].(string); ok {
		state.Output = types.StringValue(v)
	} else {
		state.Output = types.StringValue("ACCEPT")
	}
	if v, ok := data["forward"].(string); ok {
		state.Forward = types.StringValue(v)
	} else {
		state.Forward = types.StringValue("REJECT")
	}
	if v, ok := data["masq"].(string); ok {
		state.Masq = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["masq_src"].(string); ok {
		state.MasqSrc = types.StringValue(v)
	}
	if v, ok := data["masq_dest"].(string); ok {
		state.MasqDest = types.StringValue(v)
	}
	if v, ok := data["mtu_fix"].(string); ok {
		state.MtuFix = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["network"].(string); ok && v != "" {
		networks := strings.Split(v, " ")
		state.Network, _ = types.ListValueFrom(ctx, types.StringType, networks)
	}
}
