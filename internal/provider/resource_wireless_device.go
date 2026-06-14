package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewWirelessDeviceResource() resource.Resource {
	return &wirelessDeviceResource{}
}

type wirelessDeviceResource struct {
	client *JsonRpcClient
}

type wirelessDeviceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Path     types.String `tfsdk:"path"`
	Band     types.String `tfsdk:"band"`
	Channel  types.Int64  `tfsdk:"channel"`
	HTMode   types.String `tfsdk:"htmode"`
	HWMode   types.String `tfsdk:"hwmode"`
	Country  types.String `tfsdk:"country"`
	Disabled types.Bool   `tfsdk:"disabled"`
}

func (r *wirelessDeviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wireless_device"
}

func (r *wirelessDeviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an OpenWrt wireless radio device.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: wireless/<device_name>.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the wireless device (e.g., 'radio0', 'radio1').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("mac80211"),
				Description: "Wireless driver type ('mac80211', 'broadcom', etc.).",
			},
			"path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Hardware path to the wireless device.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"band": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Radio band: '2g', '5g', '6g'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"channel": schema.Int64Attribute{
				Optional:    true,
				Description: "Wireless channel number.",
			},
			"htmode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Channel width/HT mode (e.g., 'VHT80', 'HT40', 'HE80').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hwmode": schema.StringAttribute{
				Optional:    true,
				Description: "Hardware mode: '11a', '11g', '11n', '11ac', '11ax'. Auto-detected if not set.",
			},
			"country": schema.StringAttribute{
				Optional:    true,
				Description: "Country code for regulatory compliance.",
			},
			"disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Disable the radio.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *wirelessDeviceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *wirelessDeviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan wirelessDeviceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(plan)

	secName, err := r.client.UCISection(ctx, "wireless", "wifi-device", name, options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating wireless device", err.Error())
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
	if err := r.client.WifiReload(ctx); err != nil {
		tflog.Warn(ctx, "WiFi reload failed", map[string]interface{}{"error": err.Error()})
	}

	devices, err := r.client.UCIForeach(ctx, "wireless", "wifi-device")
	if err != nil {
		resp.Diagnostics.AddError("Error reading wireless device after create", err.Error())
		return
	}
	for _, dev := range devices {
		if dev[".name"] == name {
			r.optionsToModel(dev, &plan)
			break
		}
	}

	plan.ID = types.StringValue(fmt.Sprintf("wireless/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *wirelessDeviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state wirelessDeviceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	devices, err := r.client.UCIForeach(ctx, "wireless", "wifi-device")
	if err != nil {
		resp.Diagnostics.AddError("Error reading wireless devices", err.Error())
		return
	}

	var data map[string]interface{}
	for _, dev := range devices {
		if dev[".name"] == name {
			data = dev
			break
		}
	}

	if data == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(data, &state)
	state.ID = types.StringValue(fmt.Sprintf("wireless/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *wirelessDeviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan wirelessDeviceModel
	var state wirelessDeviceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(plan)

	if err := r.client.UCITSet(ctx, "wireless", name, options); err != nil {
		resp.Diagnostics.AddError("Error updating wireless device", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "wireless"); err != nil {
		resp.Diagnostics.AddError("Error committing wireless config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}
	if err := r.client.WifiReload(ctx); err != nil {
		tflog.Warn(ctx, "WiFi reload failed", map[string]interface{}{"error": err.Error()})
	}

	devices, err := r.client.UCIForeach(ctx, "wireless", "wifi-device")
	if err != nil {
		resp.Diagnostics.AddError("Error reading wireless device after update", err.Error())
		return
	}
	for _, dev := range devices {
		if dev[".name"] == name {
			r.optionsToModel(dev, &plan)
			break
		}
	}

	plan.ID = types.StringValue(fmt.Sprintf("wireless/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *wirelessDeviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state wirelessDeviceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	if err := r.client.UCIDelete(ctx, "wireless", name); err != nil {
		resp.Diagnostics.AddError("Error deleting wireless device", err.Error())
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

func (r *wirelessDeviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] != "wireless" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: wireless/<device_name>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}

func (r *wirelessDeviceResource) modelToOptions(plan wirelessDeviceModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.Type.IsNull() {
		options["type"] = plan.Type.ValueString()
	}
	if !plan.Path.IsNull() && !plan.Path.IsUnknown() {
		options["path"] = plan.Path.ValueString()
	}
	if !plan.Band.IsNull() && !plan.Band.IsUnknown() {
		options["band"] = plan.Band.ValueString()
	}
	if !plan.Channel.IsNull() {
		options["channel"] = plan.Channel.ValueInt64()
	}
	if !plan.HTMode.IsNull() && !plan.HTMode.IsUnknown() {
		options["htmode"] = plan.HTMode.ValueString()
	}
	if !plan.HWMode.IsNull() {
		options["hwmode"] = plan.HWMode.ValueString()
	}
	if !plan.Country.IsNull() {
		options["country"] = plan.Country.ValueString()
	}
	if !plan.Disabled.IsNull() && !plan.Disabled.IsUnknown() {
		if plan.Disabled.ValueBool() {
			options["disabled"] = "1"
		}
	}

	return options
}

func (r *wirelessDeviceResource) optionsToModel(data map[string]interface{}, state *wirelessDeviceModel) {
	if v, ok := data["type"].(string); ok {
		state.Type = types.StringValue(v)
	}
	if v, ok := data["path"].(string); ok {
		state.Path = types.StringValue(v)
	} else if state.Path.IsUnknown() {
		state.Path = types.StringValue("")
	}
	if v, ok := data["band"].(string); ok {
		state.Band = types.StringValue(v)
	} else if state.Band.IsUnknown() {
		state.Band = types.StringValue("")
	}
	if v, ok := data["channel"]; ok {
		if f, ok := v.(float64); ok {
			state.Channel = types.Int64Value(int64(f))
		}
	}
	if v, ok := data["htmode"].(string); ok {
		state.HTMode = types.StringValue(v)
	} else if state.HTMode.IsUnknown() {
		state.HTMode = types.StringValue("")
	}
	if v, ok := data["hwmode"].(string); ok {
		state.HWMode = types.StringValue(v)
	}
	if v, ok := data["country"].(string); ok {
		state.Country = types.StringValue(v)
	}
	if v, ok := data["disabled"].(string); ok {
		state.Disabled = types.BoolValue(v == "1" || v == "true")
	} else if state.Disabled.IsUnknown() {
		state.Disabled = types.BoolValue(false)
	}
}
