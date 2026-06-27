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

func NewSystemLEDResource() resource.Resource {
	return &systemLEDResource{}
}

type systemLEDResource struct {
	client *JsonRpcClient
}

type systemLEDModel struct {
	ID      types.String `tfsdk:"id"`
	Section types.String `tfsdk:"section"`
	Name    types.String `tfsdk:"name"`
	SysFS   types.String `tfsdk:"sysfs"`
	Trigger types.String `tfsdk:"trigger"`
	Mode    types.String `tfsdk:"mode"`
	Dev     types.String `tfsdk:"dev"`
	Default types.Bool   `tfsdk:"default"`
}

func (r *systemLEDResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_led"
}

func (r *systemLEDResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages OpenWrt system LED configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: system/<led_name>.",
			},
			"section": schema.StringAttribute{
				Computed:    true,
				Description: "Internal UCI section identifier (e.g. 'cfg0abc12') of the anonymous section backing this resource. Managed automatically; resolved on import.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "LED name displayed in LuCI.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sysfs": schema.StringAttribute{
				Required:    true,
				Description: "LED device path (e.g., 'apu:green:3').",
			},
			"trigger": schema.StringAttribute{
				Optional:    true,
				Description: "Trigger type: 'none', 'netdev', 'timer', 'heartbeat', 'gpio', 'usbdev'.",
			},
			"mode": schema.StringAttribute{
				Optional:    true,
				Description: "Trigger mode (e.g., 'link tx rx' for netdev).",
			},
			"dev": schema.StringAttribute{
				Optional:    true,
				Description: "Network device for netdev trigger.",
			},
			"default": schema.BoolAttribute{
				Optional:    true,
				Description: "LED enabled by default.",
			},
		},
	}
}

func (r *systemLEDResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *systemLEDResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan systemLEDModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(plan)

	secName, err := r.client.UCIAdd(ctx, "system", "led")
	if err != nil {
		resp.Diagnostics.AddError("Error creating LED config", err.Error())
		return
	}
	plan.Section = types.StringValue(secName)

	for key, value := range options {
		if err := r.client.UCISet(ctx, "system", secName, key, value); err != nil {
			resp.Diagnostics.AddError("Error setting LED option", err.Error())
			return
		}
	}

	if err := r.client.UCISet(ctx, "system", secName, "name", name); err != nil {
		resp.Diagnostics.AddError("Error setting LED name", err.Error())
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

func (r *systemLEDResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state systemLEDModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	data, sectionName, found, err := r.client.UCIResolveSection(ctx, "system", "led", state.Section.ValueString(), map[string]string{"name": name})
	if err != nil {
		resp.Diagnostics.AddError("Error reading LED configs", err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(data, &state)
	state.Section = types.StringValue(sectionName)
	state.ID = types.StringValue(fmt.Sprintf("system/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *systemLEDResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan systemLEDModel
	var state systemLEDModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(plan)

	_, secName, found, err := r.client.UCIResolveSection(ctx, "system", "led", state.Section.ValueString(), map[string]string{"name": name})
	if err != nil {
		resp.Diagnostics.AddError("Error reading LED configs", err.Error())
		return
	}
	if !found {
		resp.Diagnostics.AddError("Error finding LED config", "LED not found")
		return
	}

	for key, value := range options {
		if err := r.client.UCISet(ctx, "system", secName, key, value); err != nil {
			resp.Diagnostics.AddError("Error updating LED option", err.Error())
			return
		}
	}

	if err := r.client.UCICommit(ctx, "system"); err != nil {
		resp.Diagnostics.AddError("Error committing system config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.Section = types.StringValue(secName)
	plan.ID = types.StringValue(fmt.Sprintf("system/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *systemLEDResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state systemLEDModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	_, secName, found, err := r.client.UCIResolveSection(ctx, "system", "led", state.Section.ValueString(), map[string]string{"name": name})
	if err != nil {
		resp.Diagnostics.AddError("Error reading LED configs", err.Error())
		return
	}
	if !found {
		// Section already absent - treat as already deleted.
		resp.State.RemoveResource(ctx)
		return
	}

	if err := r.client.UCIDelete(ctx, "system", secName); err != nil {
		resp.Diagnostics.AddError("Error deleting LED config", err.Error())
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

func (r *systemLEDResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] != "system" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: system/<led_name>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}

func (r *systemLEDResource) modelToOptions(plan systemLEDModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.SysFS.IsNull() {
		options["sysfs"] = plan.SysFS.ValueString()
	}
	if !plan.Trigger.IsNull() {
		options["trigger"] = plan.Trigger.ValueString()
	}
	if !plan.Mode.IsNull() {
		options["mode"] = plan.Mode.ValueString()
	}
	if !plan.Dev.IsNull() {
		options["dev"] = plan.Dev.ValueString()
	}
	if !plan.Default.IsNull() {
		options["default"] = boolToString(plan.Default.ValueBool())
	}

	return options
}

func (r *systemLEDResource) optionsToModel(data map[string]interface{}, state *systemLEDModel) {
	if v, ok := data["sysfs"].(string); ok {
		state.SysFS = types.StringValue(v)
	}
	if v, ok := data["trigger"].(string); ok {
		state.Trigger = types.StringValue(v)
	}
	if v, ok := data["mode"].(string); ok {
		state.Mode = types.StringValue(v)
	}
	if v, ok := data["dev"].(string); ok {
		state.Dev = types.StringValue(v)
	}
	if v, ok := data["default"].(string); ok {
		state.Default = types.BoolValue(v == "1" || v == "true")
	}
}
