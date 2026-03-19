package resource

import (
	"context"
	"fmt"

	"github.com/bucksbunni/terraform-provider-openwrt/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &uciSectionResource{}
var _ resource.ResourceWithImportState = &uciSectionResource{}

func NewUCISectionResource() resource.Resource {
	return &uciSectionResource{}
}

type uciSectionResource struct {
	client *client.JsonRpcClient
}

type uciSectionModel struct {
	ID      types.String `tfsdk:"id"`
	Config  types.String `tfsdk:"config"`
	Type    types.String `tfsdk:"type"`
	Name    types.String `tfsdk:"name"`
	Options types.Map    `tfsdk:"options"` // map[string]string
}

func (r *uciSectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_uci_section"
}

func (r *uciSectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenWrt UCI section and its key/value options.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: <config>/<section_name>.",
			},
			"config": rschema.StringAttribute{
				Required:    true,
				Description: "Name of the UCI config file (without /etc/config/ prefix), e.g. 'network'.",
			},
			"type": rschema.StringAttribute{
				Required:    true,
				Description: "UCI section type, e.g. 'interface'.",
			},
			"name": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UCI section name. If omitted, an anonymous section is created and its name is returned.",
			},
			"options": rschema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     mapdefault.StaticValue(types.MapNull(types.StringType)),
				Description: "Map of UCI options for this section.",
			},
		},
	}
}

func (r *uciSectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.JsonRpcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.JsonRpcClient, got %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *uciSectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan uciSectionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := plan.Config.ValueString()
	typ := plan.Type.ValueString()

	options := map[string]interface{}{}
	if !plan.Options.IsNull() {
		var optMap map[string]string
		resp.Diagnostics.Append(plan.Options.ElementsAs(ctx, &optMap, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for k, v := range optMap {
			options[k] = v
		}
	}

	name := plan.Name.ValueString()

	secName, err := r.client.UCISection(ctx, config, typ, name, options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating UCI section", err.Error())
		return
	}

	if name == "" {
		name = secName
	}

	if err := r.client.UCICommit(ctx, config); err != nil {
		resp.Diagnostics.AddError("Error committing UCI config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.Name = types.StringValue(name)
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", config, name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *uciSectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state uciSectionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := state.Config.ValueString()
	name := state.Name.ValueString()

	secData, err := r.client.UCIGetAll(ctx, config, name)
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if len(secData) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	options := make(map[string]string)
	for k, v := range secData {
		if k == ".name" || k == ".type" {
			continue
		}
		switch val := v.(type) {
		case string:
			options[k] = val
		case bool:
			if val {
				options[k] = "1"
			} else {
				options[k] = "0"
			}
		case float64:
			options[k] = fmt.Sprintf("%.0f", val)
		}
	}

	optionsMap, diag := types.MapValueFrom(ctx, types.StringType, options)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Options = optionsMap
	state.ID = types.StringValue(fmt.Sprintf("%s/%s", config, name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *uciSectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan uciSectionModel
	var state uciSectionModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := plan.Config.ValueString()
	name := plan.Name.ValueString()

	options := map[string]interface{}{}
	if !plan.Options.IsNull() {
		var optMap map[string]string
		resp.Diagnostics.Append(plan.Options.ElementsAs(ctx, &optMap, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for k, v := range optMap {
			options[k] = v
		}
	}

	if err := r.client.UCITSet(ctx, config, name, options); err != nil {
		resp.Diagnostics.AddError("Error updating UCI section", err.Error())
		return
	}
	if err := r.client.UCICommit(ctx, config); err != nil {
		resp.Diagnostics.AddError("Error committing UCI config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *uciSectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state uciSectionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := state.Config.ValueString()
	name := state.Name.ValueString()

	if err := r.client.UCIDelete(ctx, config, name); err != nil {
		resp.Diagnostics.AddError("Error deleting UCI section", err.Error())
		return
	}
	if err := r.client.UCICommit(ctx, config); err != nil {
		resp.Diagnostics.AddError("Error committing UCI config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.State.RemoveResource(ctx)
}

func (r *uciSectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// ID format: config/name
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	parts := path.Root("id")
	_ = parts

	// Simple split:
	var config, name string
	_, err := fmt.Sscanf(req.ID, "%s/%s", &config, &name)
	if err != nil || config == "" || name == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected format: <config>/<section_name>",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config"), config)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
}
