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

func NewDHCPHostResource() resource.Resource {
	return &dhcpHostResource{}
}

type dhcpHostResource struct {
	client *JsonRpcClient
}

type dhcpHostModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	IP        types.String `tfsdk:"ip"`
	MAC       types.List   `tfsdk:"mac"`
	Leasetime types.String `tfsdk:"leasetime"`
}

func (r *dhcpHostResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_host"
}

func (r *dhcpHostResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a static DHCP host reservation in OpenWrt.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: dhcp/<host_name>.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Hostname for the DHCP reservation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip": schema.StringAttribute{
				Required:    true,
				Description: "IP address to assign (e.g., '192.168.1.100').",
			},
			"mac": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "MAC addresses to match (e.g., ['00:11:22:33:44:55']).",
			},
			"leasetime": schema.StringAttribute{
				Optional:    true,
				Description: "Lease time override (e.g., '12h', 'infinite').",
			},
		},
	}
}

func (r *dhcpHostResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dhcpHostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dhcpHostModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(ctx, plan)

	_, err := r.client.UCISection(ctx, "dhcp", "host", name, options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating DHCP host", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "dhcp"); err != nil {
		resp.Diagnostics.AddError("Error committing DHCP config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue(fmt.Sprintf("dhcp/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dhcpHostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dhcpHostModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	data, err := r.client.UCIGetAll(ctx, "dhcp", name)
	if err != nil || len(data) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(ctx, data, &state)
	state.ID = types.StringValue(fmt.Sprintf("dhcp/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dhcpHostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dhcpHostModel
	var state dhcpHostModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(ctx, plan)

	if err := r.client.UCITSet(ctx, "dhcp", name, options); err != nil {
		resp.Diagnostics.AddError("Error updating DHCP host", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "dhcp"); err != nil {
		resp.Diagnostics.AddError("Error committing DHCP config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dhcpHostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dhcpHostModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	if err := r.client.UCIDelete(ctx, "dhcp", name); err != nil {
		resp.Diagnostics.AddError("Error deleting DHCP host", err.Error())
		return
	}
	if err := r.client.UCICommit(ctx, "dhcp"); err != nil {
		resp.Diagnostics.AddError("Error committing DHCP config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.State.RemoveResource(ctx)
}

func (r *dhcpHostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] != "dhcp" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: dhcp/<host_name>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}

func (r *dhcpHostResource) modelToOptions(ctx context.Context, plan dhcpHostModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.IP.IsNull() {
		options["ip"] = plan.IP.ValueString()
	}
	if !plan.MAC.IsNull() {
		var macList []string
		plan.MAC.ElementsAs(ctx, &macList, false)
		options["mac"] = strings.Join(macList, " ")
	}
	if !plan.Leasetime.IsNull() {
		options["leasetime"] = plan.Leasetime.ValueString()
	}

	return options
}

func (r *dhcpHostResource) optionsToModel(ctx context.Context, data map[string]interface{}, state *dhcpHostModel) {
	if v, ok := data["name"].(string); ok {
		state.Name = types.StringValue(v)
	}
	if v, ok := data["ip"].(string); ok {
		state.IP = types.StringValue(v)
	}
	if v, ok := data["mac"].(string); ok && v != "" {
		macList := strings.Split(v, " ")
		state.MAC, _ = types.ListValueFrom(ctx, types.StringType, macList)
	}
	if v, ok := data["leasetime"].(string); ok {
		state.Leasetime = types.StringValue(v)
	}
}
