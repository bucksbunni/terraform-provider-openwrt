package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewSysRebootResource() resource.Resource {
	return &sysRebootResource{}
}

type sysRebootResource struct {
	client *JsonRpcClient
}

type sysRebootModel struct {
	ID      types.String `tfsdk:"id"`
	Delay   types.Int64  `tfsdk:"delay"`
	Message types.String `tfsdk:"message"`
}

func (r *sysRebootResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sys_reboot"
}

func (r *sysRebootResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Triggers a reboot of the OpenWrt device.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: sys_reboot.",
			},
			"delay": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Description: "Delay in seconds before rebooting.",
			},
			"message": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Rebooting via Terraform"),
				Description: "Shutdown message.",
			},
		},
	}
}

func (r *sysRebootResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *sysRebootResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sysRebootModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	delay := plan.Delay.ValueInt64()
	message := plan.Message.ValueString()

	if delay > 0 {
		tflog.Info(ctx, fmt.Sprintf("Waiting %d seconds before reboot...", delay))
		time.Sleep(time.Duration(delay) * time.Second)
	}

	tflog.Info(ctx, "Rebooting OpenWrt device...")
	result, err := r.client.SysCall(ctx, "exec", map[string]interface{}{
		"command": fmt.Sprintf("/sbin/shutdown -r %s '%s'", "-t 1", message),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error triggering reboot", err.Error())
		return
	}

	var resultMap map[string]interface{}
	if err := json.Unmarshal(result, &resultMap); err != nil {
		tflog.Warn(ctx, "Could not parse reboot response", map[string]interface{}{"error": err.Error()})
	}

	tflog.Info(ctx, "Reboot command sent successfully", map[string]interface{}{"result": string(result)})

	plan.ID = types.StringValue("sys_reboot")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sysRebootResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sysRebootModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue("sys_reboot")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *sysRebootResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan sysRebootModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	delay := plan.Delay.ValueInt64()
	message := plan.Message.ValueString()

	if delay > 0 {
		tflog.Info(ctx, fmt.Sprintf("Waiting %d seconds before reboot...", delay))
		time.Sleep(time.Duration(delay) * time.Second)
	}

	tflog.Info(ctx, "Rebooting OpenWrt device...")

	result, err := r.client.SysCall(ctx, "exec", map[string]interface{}{
		"command": fmt.Sprintf("/sbin/shutdown -r %s '%s'", "-t 1", message),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error triggering reboot", err.Error())
		return
	}

	tflog.Info(ctx, "Reboot command sent successfully", map[string]interface{}{"result": string(result)})

	plan.ID = types.StringValue("sys_reboot")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sysRebootResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sysRebootModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddWarning("Reboot cannot be undone", "The device has been rebooted. This resource will be removed from state.")
	resp.State.RemoveResource(ctx)
}

func (r *sysRebootResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "sys_reboot")...)
}