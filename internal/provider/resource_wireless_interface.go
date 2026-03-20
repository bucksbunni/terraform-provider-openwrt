package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewWirelessInterfaceResource() resource.Resource {
	return &wirelessInterfaceResource{}
}

type wirelessInterfaceResource struct {
	client *JsonRpcClient
}

type wirelessInterfaceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Device     types.String `tfsdk:"device"`
	Mode       types.String `tfsdk:"mode"`
	SSID       types.String `tfsdk:"ssid"`
	Encryption types.String `tfsdk:"encryption"`
	Key        types.String `tfsdk:"key"`
	Network    types.String `tfsdk:"network"`
	Disabled   types.Bool   `tfsdk:"disabled"`
	Hidden     types.Bool   `tfsdk:"hidden"`
	MACFilter  types.String `tfsdk:"macfilter"`
	MACList    types.String `tfsdk:"maclist"`
	Isolate    types.Bool   `tfsdk:"isolate"`
}

func (r *wirelessInterfaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wireless_iface"
}

func (r *wirelessInterfaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an OpenWrt wireless interface (WiFi SSID).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: wireless/<iface_name>.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the wireless interface (e.g., 'wifinet0', 'wifinet1').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"device": schema.StringAttribute{
				Required:    true,
				Description: "Radio device to use (e.g., 'radio0').",
			},
			"mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ap"),
				Description: "Operation mode: 'ap', 'sta', 'adhoc', 'monitor', 'wds'.",
			},
			"ssid": schema.StringAttribute{
				Required:    true,
				Description: "Wireless network name (SSID).",
			},
			"encryption": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("psk2"),
				Description: "Encryption mode: 'psk', 'psk2', 'psk2+ccmp', 'sae', 'wpa3', etc.",
			},
			"key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Wireless passphrase/PSK.",
			},
			"network": schema.StringAttribute{
				Optional:    true,
				Description: "Network to attach (e.g., 'lan', 'guest').",
			},
			"disabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Disable this wireless interface.",
			},
			"hidden": schema.BoolAttribute{
				Optional:    true,
				Description: "Hide SSID from broadcast.",
			},
			"macfilter": schema.StringAttribute{
				Optional:    true,
				Description: "MAC filter mode: 'disable', 'allow', 'deny'.",
			},
			"maclist": schema.StringAttribute{
				Optional:    true,
				Description: "MAC address list for filtering, space-separated.",
			},
			"isolate": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable client isolation.",
			},
		},
	}
}

func (r *wirelessInterfaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *wirelessInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan wirelessInterfaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(plan)

	secName, err := r.client.UCISection(ctx, "wireless", "wifi-iface", name, options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating wireless interface", err.Error())
		return
	}

	if name == "" {
		name = secName
	}

	if err := r.client.UCICommit(ctx, "wireless"); err != nil {
		resp.Diagnostics.AddError("Error committing wireless config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue(fmt.Sprintf("wireless/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *wirelessInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state wirelessInterfaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	data, err := r.client.UCIGetAll(ctx, "wireless", name)
	if err != nil || len(data) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(data, &state)
	state.ID = types.StringValue(fmt.Sprintf("wireless/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *wirelessInterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan wirelessInterfaceModel
	var state wirelessInterfaceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(plan)

	if err := r.client.UCITSet(ctx, "wireless", name, options); err != nil {
		resp.Diagnostics.AddError("Error updating wireless interface", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "wireless"); err != nil {
		resp.Diagnostics.AddError("Error committing wireless config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *wirelessInterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state wirelessInterfaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	if err := r.client.UCIDelete(ctx, "wireless", name); err != nil {
		resp.Diagnostics.AddError("Error deleting wireless interface", err.Error())
		return
	}
	if err := r.client.UCICommit(ctx, "wireless"); err != nil {
		resp.Diagnostics.AddError("Error committing wireless config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.State.RemoveResource(ctx)
}

func (r *wirelessInterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] != "wireless" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: wireless/<iface_name>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}

func (r *wirelessInterfaceResource) modelToOptions(plan wirelessInterfaceModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.Device.IsNull() {
		options["device"] = plan.Device.ValueString()
	}
	if !plan.Mode.IsNull() {
		options["mode"] = plan.Mode.ValueString()
	}
	if !plan.SSID.IsNull() {
		options["ssid"] = plan.SSID.ValueString()
	}
	if !plan.Encryption.IsNull() {
		options["encryption"] = plan.Encryption.ValueString()
	}
	if !plan.Key.IsNull() {
		options["key"] = plan.Key.ValueString()
	}
	if !plan.Network.IsNull() {
		options["network"] = plan.Network.ValueString()
	}
	if !plan.Disabled.IsNull() {
		if plan.Disabled.ValueBool() {
			options["disabled"] = "1"
		}
	}
	if !plan.Hidden.IsNull() {
		if plan.Hidden.ValueBool() {
			options["hidden"] = "1"
		}
	}
	if !plan.MACFilter.IsNull() {
		options["macfilter"] = plan.MACFilter.ValueString()
	}
	if !plan.MACList.IsNull() {
		options["maclist"] = plan.MACList.ValueString()
	}
	if !plan.Isolate.IsNull() {
		if plan.Isolate.ValueBool() {
			options["isolate"] = "1"
		}
	}

	return options
}

func (r *wirelessInterfaceResource) optionsToModel(data map[string]interface{}, state *wirelessInterfaceModel) {
	if v, ok := data["device"].(string); ok {
		state.Device = types.StringValue(v)
	}
	if v, ok := data["mode"].(string); ok {
		state.Mode = types.StringValue(v)
	}
	if v, ok := data["ssid"].(string); ok {
		state.SSID = types.StringValue(v)
	}
	if v, ok := data["encryption"].(string); ok {
		state.Encryption = types.StringValue(v)
	}
	if v, ok := data["key"].(string); ok {
		state.Key = types.StringValue(v)
	}
	if v, ok := data["network"].(string); ok {
		state.Network = types.StringValue(v)
	}
	if v, ok := data["disabled"].(string); ok {
		state.Disabled = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["hidden"].(string); ok {
		state.Hidden = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["macfilter"].(string); ok {
		state.MACFilter = types.StringValue(v)
	}
	if v, ok := data["maclist"].(string); ok {
		state.MACList = types.StringValue(v)
	}
	if v, ok := data["isolate"].(string); ok {
		state.Isolate = types.BoolValue(v == "1" || v == "true")
	}
}
