package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewSysModulesResource() resource.Resource {
	return &sysModulesResource{}
}

type sysModulesResource struct {
	client *JsonRpcClient
}

type sysModulesModel struct {
	ID      types.String   `tfsdk:"id"`
	Modules types.List    `tfsdk:"modules"`
}

func (r *sysModulesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sys_modules"
}

func (r *sysModulesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages kernel modules to load at boot time on OpenWrt device. Creates entries in /etc/modules.d/.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: sys_modules.",
			},
			"modules": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of kernel module names to load at boot (e.g., ['ath10k_pci', 'ath10k_core']).",
			},
		},
	}
}

func (r *sysModulesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *sysModulesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sysModulesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var modules []string
	diag := plan.Modules.ElementsAs(ctx, &modules, false)
	if diag.HasError() {
		return
	}

	if err := r.writeModulesFile(ctx, modules); err != nil {
		resp.Diagnostics.AddError("Error writing modules configuration", err.Error())
		return
	}

	plan.ID = types.StringValue("sys_modules")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sysModulesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sysModulesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	modules, err := r.readModulesFile(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading modules configuration", err.Error())
		return
	}

	if len(modules) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	modulesValue, diag := types.ListValueFrom(ctx, types.StringType, modules)
	if diag.HasError() {
		return
	}
	state.Modules = modulesValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *sysModulesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan sysModulesModel
	var state sysModulesModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var modules []string
	diag := plan.Modules.ElementsAs(ctx, &modules, false)
	if diag.HasError() {
		return
	}

	if err := r.writeModulesFile(ctx, modules); err != nil {
		resp.Diagnostics.AddError("Error writing modules configuration", err.Error())
		return
	}

	plan.ID = types.StringValue("sys_modules")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sysModulesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sysModulesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.writeModulesFile(ctx, []string{}); err != nil {
		resp.Diagnostics.AddWarning("Error clearing modules configuration", err.Error())
	}

	resp.State.RemoveResource(ctx)
}

func (r *sysModulesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "sys_modules")...)
}

func (r *sysModulesResource) readModulesFile(ctx context.Context) ([]string, error) {
	modulesDir := "/etc/modules.d"

	result, err := r.client.SysCall(ctx, "exec", fmt.Sprintf("ls -la %s/ 2>/dev/null", modulesDir))
	if err != nil {
		return nil, err
	}

	var resultMap map[string]interface{}
	if err := json.Unmarshal(result, &resultMap); err != nil {
		return nil, err
	}

	msg, _ := resultMap["msg"].(string)
	if msg == "" || strings.Contains(msg, "No such file") {
		return []string{}, nil
	}

	listResult, err := r.client.SysCall(ctx, "exec", fmt.Sprintf("cat %s/* 2>/dev/null", modulesDir))
	if err != nil {
		return nil, err
	}

	var listMap map[string]interface{}
	if err := json.Unmarshal(listResult, &listMap); err != nil {
		return nil, err
	}

	listMsg, _ := listMap["msg"].(string)
	if listMsg == "" {
		return []string{}, nil
	}

	var modules []string
	lines := strings.Split(listMsg, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			modules = append(modules, line)
		}
	}

	return modules, nil
}

func (r *sysModulesResource) writeModulesFile(ctx context.Context, modules []string) error {
	modulesDir := "/etc/modules.d"

	sort.Strings(modules)

	seen := make(map[string]bool)
	var uniqueModules []string
	for _, m := range modules {
		if !seen[m] {
			seen[m] = true
			uniqueModules = append(uniqueModules, m)
		}
	}

	command := fmt.Sprintf("mkdir -p %s", modulesDir)
	if _, err := r.client.SysCall(ctx, "exec", command); err != nil {
		return fmt.Errorf("failed to create modules directory: %w", err)
	}

	existingModules, err := r.readModulesFile(ctx)
	if err != nil {
		tflog.Warn(ctx, "Could not read existing modules, will overwrite", map[string]interface{}{"error": err.Error()})
		existingModules = []string{}
	}

	existingMap := make(map[string]bool)
	for _, m := range existingModules {
		existingMap[m] = true
	}

	for _, m := range uniqueModules {
		if existingMap[m] {
			continue
		}

		filename := filepath.Join(modulesDir, m)
		command = fmt.Sprintf("echo '%s' > %s", m, filename)
		if _, err := r.client.SysCall(ctx, "exec", command); err != nil {
			return fmt.Errorf("failed to write module %s: %w", m, err)
		}
		tflog.Info(ctx, fmt.Sprintf("Added module to boot load: %s", m))
	}

	modulesToRemove := make(map[string]bool)
	for _, m := range existingModules {
		modulesToRemove[m] = true
	}
	for _, m := range uniqueModules {
		delete(modulesToRemove, m)
	}

	for m := range modulesToRemove {
		filename := filepath.Join(modulesDir, m)
		command = fmt.Sprintf("rm -f %s", filename)
		if _, err := r.client.SysCall(ctx, "exec", command); err != nil {
			tflog.Warn(ctx, fmt.Sprintf("Failed to remove module %s: %v", m, err))
		}
		tflog.Info(ctx, fmt.Sprintf("Removed module from boot load: %s", m))
	}

	return nil
}