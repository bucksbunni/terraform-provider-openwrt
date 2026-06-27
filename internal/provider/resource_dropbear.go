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

func NewDropbearResource() resource.Resource {
	return &dropbearResource{}
}

type dropbearResource struct {
	client *JsonRpcClient
}

type dropbearModel struct {
	ID               types.String `tfsdk:"id"`
	Section          types.String `tfsdk:"section"`
	Name             types.String `tfsdk:"name"`
	PasswordAuth     types.Bool   `tfsdk:"password_auth"`
	Port             types.Int64  `tfsdk:"port"`
	RootPasswordAuth types.Bool   `tfsdk:"root_password_auth"`
	RootLogin        types.Bool   `tfsdk:"root_login"`
}

func (r *dropbearResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dropbear"
}

func (r *dropbearResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages OpenWrt Dropbear SSH server settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: dropbear/<instance_name>.",
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
				Description: "Instance name (e.g., 'main').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password_auth": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable password authentication.",
			},
			"port": schema.Int64Attribute{
				Optional:    true,
				Description: "SSH listen port (default: 22).",
			},
			"root_password_auth": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable root password authentication.",
			},
			"root_login": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable root login.",
			},
		},
	}
}

func (r *dropbearResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dropbearResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dropbearModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(plan)

	secName, err := r.client.UCIAdd(ctx, "dropbear", "dropbear")
	if err != nil {
		resp.Diagnostics.AddError("Error creating dropbear config", err.Error())
		return
	}
	plan.Section = types.StringValue(secName)

	for key, value := range options {
		if err := r.client.UCISet(ctx, "dropbear", secName, key, value); err != nil {
			resp.Diagnostics.AddError("Error setting dropbear option", err.Error())
			return
		}
	}

	if err := r.client.UCISet(ctx, "dropbear", secName, "name", name); err != nil {
		resp.Diagnostics.AddError("Error setting dropbear name", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "dropbear"); err != nil {
		resp.Diagnostics.AddError("Error committing dropbear config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue(fmt.Sprintf("dropbear/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dropbearResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dropbearModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	data, sectionName, found, err := r.client.UCIResolveSection(ctx, "dropbear", "dropbear", state.Section.ValueString(), map[string]string{"name": name})
	if err != nil {
		resp.Diagnostics.AddError("Error reading dropbear configs", err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(data, &state)
	state.Section = types.StringValue(sectionName)
	state.ID = types.StringValue(fmt.Sprintf("dropbear/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dropbearResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dropbearModel
	var state dropbearModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(plan)

	_, secName, found, err := r.client.UCIResolveSection(ctx, "dropbear", "dropbear", state.Section.ValueString(), map[string]string{"name": name})
	if err != nil {
		resp.Diagnostics.AddError("Error reading dropbear configs", err.Error())
		return
	}
	if !found {
		resp.Diagnostics.AddError("Error finding dropbear config", "instance not found")
		return
	}

	for key, value := range options {
		if err := r.client.UCISet(ctx, "dropbear", secName, key, value); err != nil {
			resp.Diagnostics.AddError("Error updating dropbear option", err.Error())
			return
		}
	}

	if err := r.client.UCICommit(ctx, "dropbear"); err != nil {
		resp.Diagnostics.AddError("Error committing dropbear config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.Section = types.StringValue(secName)
	plan.ID = types.StringValue(fmt.Sprintf("dropbear/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dropbearResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dropbearModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	_, secName, found, err := r.client.UCIResolveSection(ctx, "dropbear", "dropbear", state.Section.ValueString(), map[string]string{"name": name})
	if err != nil {
		resp.Diagnostics.AddError("Error reading dropbear configs", err.Error())
		return
	}
	if !found {
		// Section already absent - treat as already deleted.
		resp.State.RemoveResource(ctx)
		return
	}

	if err := r.client.UCIDelete(ctx, "dropbear", secName); err != nil {
		resp.Diagnostics.AddError("Error deleting dropbear config", err.Error())
		return
	}
	if err := r.client.UCICommit(ctx, "dropbear"); err != nil {
		resp.Diagnostics.AddError("Error committing dropbear config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.State.RemoveResource(ctx)
}

func (r *dropbearResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] != "dropbear" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: dropbear/<instance_name>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}

func (r *dropbearResource) modelToOptions(plan dropbearModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.PasswordAuth.IsNull() {
		options["PasswordAuth"] = boolToString(plan.PasswordAuth.ValueBool())
	}
	if !plan.Port.IsNull() {
		options["Port"] = plan.Port.ValueInt64()
	}
	if !plan.RootPasswordAuth.IsNull() {
		options["RootPasswordAuth"] = boolToString(plan.RootPasswordAuth.ValueBool())
	}
	if !plan.RootLogin.IsNull() {
		options["RootLogin"] = boolToString(plan.RootLogin.ValueBool())
	}

	return options
}

func (r *dropbearResource) optionsToModel(data map[string]interface{}, state *dropbearModel) {
	if v, ok := data["PasswordAuth"].(string); ok {
		state.PasswordAuth = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["Port"]; ok {
		if f, ok := v.(float64); ok {
			state.Port = types.Int64Value(int64(f))
		}
	}
	if v, ok := data["RootPasswordAuth"].(string); ok {
		state.RootPasswordAuth = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["RootLogin"].(string); ok {
		state.RootLogin = types.BoolValue(v == "1" || v == "true")
	}
}
