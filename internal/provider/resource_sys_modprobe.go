package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewSysModprobeResource() resource.Resource {
	return &sysModprobeResource{}
}

type sysModprobeResource struct {
	client *JsonRpcClient
}

type sysModprobeModel struct {
	ID     types.String          `tfsdk:"id"`
	Name   types.String          `tfsdk:"name"`
	Action types.String          `tfsdk:"action"`
	Param  types.Map             `tfsdk:"param"`
	Output types.String          `tfsdk:"output"`
}

func (r *sysModprobeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sys_modprobe"
}

func (r *sysModprobeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Loads or unloads a kernel module on the OpenWrt device.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: sys_modprobe/<module_name>.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the kernel module to load/unload.",
			},
			"action": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("load"),
				Description: "Action to perform: 'load' or 'unload'.",
			},
			"param": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Module parameters as key-value pairs.",
			},
			"output": schema.StringAttribute{
				Computed:    true,
				Description: "Output from the modprobe command.",
			},
		},
	}
}

func (r *sysModprobeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *sysModprobeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sysModprobeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	action := plan.Action.ValueString()

	output, err := r.executeModprobe(ctx, name, action, &plan.Param)
	if err != nil {
		resp.Diagnostics.AddError("Error executing modprobe", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("sys_modprobe/%s", name))
	plan.Output = types.StringValue(output)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sysModprobeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sysModprobeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	loaded, err := r.isModuleLoaded(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error checking module status", err.Error())
		return
	}

	if !loaded {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *sysModprobeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan sysModprobeModel
	var state sysModprobeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	action := plan.Action.ValueString()

	output, err := r.executeModprobe(ctx, name, action, &plan.Param)
	if err != nil {
		resp.Diagnostics.AddError("Error executing modprobe", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("sys_modprobe/%s", name))
	plan.Output = types.StringValue(output)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sysModprobeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sysModprobeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	tflog.Info(ctx, fmt.Sprintf("Unloading kernel module: %s", name))

	output, err := r.executeModprobe(ctx, name, "unload", nil)
	if err != nil {
		resp.Diagnostics.AddWarning("Error unloading module", err.Error())
	}

	tflog.Info(ctx, "Modprobe unload completed", map[string]interface{}{"output": output})

	resp.State.RemoveResource(ctx)
}

func (r *sysModprobeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] != "sys_modprobe" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: sys_modprobe/<module_name>")
		return
	}

	moduleName := parts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), fmt.Sprintf("sys_modprobe/%s", moduleName))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), moduleName)...)
}

func (r *sysModprobeResource) executeModprobe(ctx context.Context, name, action string, params *types.Map) (string, error) {
	var command string

	if action == "unload" {
		command = fmt.Sprintf("modprobe -r %s 2>&1", name)
	} else {
		var paramStr string
		if params != nil && !params.IsNull() && !params.IsUnknown() {
			paramMap := make(map[string]string)
			diag := params.ElementsAs(ctx, &paramMap, false)
			if diag.HasError() {
				return "", fmt.Errorf("failed to parse module parameters")
			}
			var paramsSlice []string
			for k, v := range paramMap {
				paramsSlice = append(paramsSlice, fmt.Sprintf("%s=%s", k, v))
			}
			if len(paramsSlice) > 0 {
				paramStr = " " + strings.Join(paramsSlice, " ")
			}
		}
		command = fmt.Sprintf("modprobe %s%s 2>&1", name, paramStr)
	}

	tflog.Info(ctx, fmt.Sprintf("Executing: %s", command))

	result, err := r.client.SysCall(ctx, "exec", command)
	if err != nil {
		return "", fmt.Errorf("modprobe failed: %w", err)
	}

	var resultMap map[string]interface{}
	if err := json.Unmarshal(result, &resultMap); err != nil {
		return string(result), nil
	}

	if code, ok := resultMap["code"].(float64); ok && code != 0 {
		if msg, ok := resultMap["msg"].(string); ok {
			return msg, fmt.Errorf("modprobe returned code %d: %s", int(code), msg)
		}
		return string(result), fmt.Errorf("modprobe returned code %d", int(code))
	}

	if msg, ok := resultMap["msg"].(string); ok {
		return msg, nil
	}

	return string(result), nil
}

func (r *sysModprobeResource) isModuleLoaded(ctx context.Context, name string) (bool, error) {
	result, err := r.client.SysCall(ctx, "exec", fmt.Sprintf("lsmod | grep -q ^%s && echo loaded || echo not_loaded", name))
	if err != nil {
		return false, err
	}

	var resultMap map[string]interface{}
	if err := json.Unmarshal(result, &resultMap); err != nil {
		return strings.Contains(string(result), "loaded"), nil
	}

	if msg, ok := resultMap["msg"].(string); ok {
		return strings.Contains(msg, "loaded"), nil
	}

	return false, nil
}