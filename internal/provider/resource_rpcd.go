package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewRPCDResource() resource.Resource {
	return &rpcdResource{}
}

type rpcdResource struct {
	client *JsonRpcClient
}

type rpcdModel struct {
	ID      types.String `tfsdk:"id"`
	Socket  types.String `tfsdk:"socket"`
	Timeout types.Int64  `tfsdk:"timeout"`
}

func (r *rpcdResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rpcd"
}

func (r *rpcdResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages OpenWrt RPC daemon (ubus) settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: rpcd.",
			},
			"socket": schema.StringAttribute{
				Optional:    true,
				Description: "Ubus socket path.",
			},
			"timeout": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(30),
				Description: "Session timeout in seconds.",
			},
		},
	}
}

func (r *rpcdResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *rpcdResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan rpcdModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := r.modelToOptions(plan)

	_, err := r.client.UCISection(ctx, "rpcd", "rpcd", "", options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating rpcd config", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "rpcd"); err != nil {
		resp.Diagnostics.AddError("Error committing rpcd config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue("rpcd")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *rpcdResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state rpcdModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data, err := r.client.UCIGetAll(ctx, "rpcd", "")
	if err != nil || len(data) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(data, &state)
	state.ID = types.StringValue("rpcd")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *rpcdResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan rpcdModel
	var state rpcdModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := r.modelToOptions(plan)

	if err := r.client.UCITSet(ctx, "rpcd", "", options); err != nil {
		resp.Diagnostics.AddError("Error updating rpcd config", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "rpcd"); err != nil {
		resp.Diagnostics.AddError("Error committing rpcd config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *rpcdResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Delete not supported", "RPCD settings cannot be deleted, only modified.")
}

func (r *rpcdResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	if !strings.HasPrefix(req.ID, "rpcd") {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: rpcd")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "rpcd")...)
}

func (r *rpcdResource) modelToOptions(plan rpcdModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.Socket.IsNull() {
		options["socket"] = plan.Socket.ValueString()
	}
	if !plan.Timeout.IsNull() {
		options["timeout"] = plan.Timeout.ValueInt64()
	}

	return options
}

func (r *rpcdResource) optionsToModel(data map[string]interface{}, state *rpcdModel) {
	if v, ok := data["socket"].(string); ok {
		state.Socket = types.StringValue(v)
	}
	if v, ok := data["timeout"]; ok {
		if f, ok := v.(float64); ok {
			state.Timeout = types.Int64Value(int64(f))
		}
	}
}
