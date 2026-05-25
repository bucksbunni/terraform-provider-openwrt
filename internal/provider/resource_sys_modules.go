package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewSysModulesResource() resource.Resource {
	return &sysModulesResource{}
}

// sysModulesResource manages kernel modules loaded at boot time via /etc/modules.d.
// Uses a dedicated mutex (mu) to prevent race conditions during concurrent Read operations
// when multiple resources refresh simultaneously, which can cause inconsistent module detection.
type sysModulesResource struct {
	client *JsonRpcClient
	mu     sync.Mutex
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

// Read checks the current state of kernel modules. Uses mutex to ensure
// consistent reads when multiple resources refresh concurrently.
func (r *sysModulesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	r.mu.Lock()
	defer r.mu.Unlock()

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

// readModulesFile reads kernel modules from a marker file in /etc/modules.d/.
// The sys.exec RPC can return responses in two formats: JSON-wrapped {"msg": "..."}
// or a plain string. This function handles both formats for reliable detection.
// Uses marker file "zz-terraform-managed" to isolate Terraform-managed modules.
func (r *sysModulesResource) readModulesFile(ctx context.Context) ([]string, error) {
	modulesDir := "/etc/modules.d"
	terraformMarker := "zz-terraform-managed"

	markerPath := filepath.Join(modulesDir, terraformMarker)
	result, err := r.client.SysCall(ctx, "exec", fmt.Sprintf("test -f %s && cat %s || echo ''", markerPath, markerPath))
	if err != nil {
		return nil, err
	}

	// Handle JSON-wrapped response format: {"msg": "ath10k_pci\nath10k_core\n"}
	var resultMap map[string]interface{}
	if err := json.Unmarshal(result, &resultMap); err == nil {
		if msg, ok := resultMap["msg"].(string); ok {
			return r.parseModulesFromOutput(msg), nil
		}
	}

	// Handle plain string response format: "ath10k_pci\nath10k_core\n"
	var msg string
	if err := json.Unmarshal(result, &msg); err == nil {
		return r.parseModulesFromOutput(msg), nil
	}

	return []string{}, nil
}

func (r *sysModulesResource) parseModulesFromOutput(msg string) []string {
	if msg == "" {
		return []string{}
	}
	var modules []string
	lines := strings.Split(msg, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			modules = append(modules, line)
		}
	}
	return modules
}

func (r *sysModulesResource) writeModulesFile(ctx context.Context, modules []string) error {
	modulesDir := "/etc/modules.d"
	terraformMarker := "zz-terraform-managed"

	sort.Strings(modules)

	seen := make(map[string]bool)
	var uniqueModules []string
	for _, m := range modules {
		if !seen[m] {
			seen[m] = true
			uniqueModules = append(uniqueModules, m)
		}
	}

	markerPath := filepath.Join(modulesDir, terraformMarker)

	if len(uniqueModules) == 0 {
		command := fmt.Sprintf("rm -f %s", markerPath)
		if _, err := r.client.SysCall(ctx, "exec", command); err != nil {
			tflog.Warn(ctx, fmt.Sprintf("Failed to remove marker file: %v", err))
		}
		return nil
	}

	content := strings.Join(uniqueModules, "\n")
	command := fmt.Sprintf("echo '%s' > %s", content, markerPath)
	if _, err := r.client.SysCall(ctx, "exec", command); err != nil {
		return fmt.Errorf("failed to write modules file: %w", err)
	}

	tflog.Info(ctx, fmt.Sprintf("Updated boot modules: %v", uniqueModules))

	return nil
}