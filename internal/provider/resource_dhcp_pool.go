package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewDHCPPoolResource() resource.Resource {
	return &dhcpPoolResource{}
}

type dhcpPoolResource struct {
	client *JsonRpcClient
}

type dhcpPoolModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Interface types.String `tfsdk:"interface"`
	Start     types.Int64  `tfsdk:"start"`
	Limit     types.Int64  `tfsdk:"limit"`
	Leasetime types.String `tfsdk:"leasetime"`
	DHCPv4    types.String `tfsdk:"dhcpv4"`
	DHCPv6    types.String `tfsdk:"dhcpv6"`
	RA        types.String `tfsdk:"ra"`
	RAFlags   types.List   `tfsdk:"ra_flags"`
	Ignore    types.Bool   `tfsdk:"ignore"`
}

func (r *dhcpPoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_pool"
}

func (r *dhcpPoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DHCP pool for an OpenWrt network interface.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: dhcp/<interface_name>.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the DHCP pool (typically matches the interface name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"interface": schema.StringAttribute{
				Required:    true,
				Description: "The network interface to serve DHCP on (e.g., 'lan', 'guest').",
			},
			"start": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(100),
				Description: "First IP in the DHCP range.",
			},
			"limit": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(150),
				Description: "Number of IPs to allocate (total = limit).",
			},
			"leasetime": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("12h"),
				Description: "DHCP lease time (e.g., '12h', '24h', 'infinite').",
			},
			"dhcpv4": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("server"),
				Description: "DHCPv4 mode: 'server', 'none'.",
			},
			"dhcpv6": schema.StringAttribute{
				Optional:    true,
				Description: "DHCPv6 mode: 'server', 'hybrid', 'none'.",
			},
			"ra": schema.StringAttribute{
				Optional:    true,
				Description: "Router Advertisement mode: 'server', 'hybrid', 'relay', 'disabled'.",
			},
			"ra_flags": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "RA flags (e.g., ['managed-config', 'other-config']).",
			},
			"ignore": schema.BoolAttribute{
				Optional:    true,
				Description: "Ignore this interface for DHCP (disable DHCP server).",
			},
		},
	}
}

func (r *dhcpPoolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dhcpPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dhcpPoolModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(ctx, plan)

	secName, err := r.client.UCIAdd(ctx, "dhcp", "dhcp")
	if err != nil {
		resp.Diagnostics.AddError("Error creating DHCP pool", err.Error())
		return
	}

	for key, value := range options {
		if err := r.client.UCISet(ctx, "dhcp", secName, key, value); err != nil {
			resp.Diagnostics.AddError("Error setting DHCP pool option", err.Error())
			return
		}
	}

	if err := r.client.UCISet(ctx, "dhcp", secName, "name", name); err != nil {
		resp.Diagnostics.AddError("Error setting DHCP pool name", err.Error())
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

func (r *dhcpPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dhcpPoolModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	pools, err := r.client.UCIForeach(ctx, "dhcp", "dhcp")
	if err != nil {
		resp.Diagnostics.AddError("Error reading DHCP pools", err.Error())
		return
	}

	var data map[string]interface{}
	for _, pool := range pools {
		nameVal, ok := pool["name"].(string)
		if ok && nameVal == name {
			data = pool
			break
		}
	}

	if data == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(data, &state)
	state.ID = types.StringValue(fmt.Sprintf("dhcp/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dhcpPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dhcpPoolModel
	var state dhcpPoolModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(ctx, plan)

	pools, err := r.client.UCIForeach(ctx, "dhcp", "dhcp")
	if err != nil {
		resp.Diagnostics.AddError("Error reading DHCP pools", err.Error())
		return
	}

	var secName string
	for _, p := range pools {
		if p["name"] == name {
			secName = p[".name"].(string)
			break
		}
	}

	if secName == "" {
		resp.Diagnostics.AddError("Error finding DHCP pool", "pool not found")
		return
	}

	for key, value := range options {
		if err := r.client.UCISet(ctx, "dhcp", secName, key, value); err != nil {
			resp.Diagnostics.AddError("Error updating DHCP pool option", err.Error())
			return
		}
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

func (r *dhcpPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dhcpPoolModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	pools, err := r.client.UCIForeach(ctx, "dhcp", "dhcp")
	if err != nil {
		resp.Diagnostics.AddError("Error reading DHCP pools", err.Error())
		return
	}

	var secName string
	for _, p := range pools {
		if p["name"] == name {
			secName = p[".name"].(string)
			break
		}
	}

	if secName != "" {
		if err := r.client.UCIDelete(ctx, "dhcp", secName); err != nil {
			resp.Diagnostics.AddError("Error deleting DHCP pool", err.Error())
			return
		}
		if err := r.client.UCICommit(ctx, "dhcp"); err != nil {
			resp.Diagnostics.AddError("Error committing DHCP config", err.Error())
			return
		}
		if err := r.client.UCIApply(ctx, false); err != nil {
			tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
		}
	}

	resp.State.RemoveResource(ctx)
}

func (r *dhcpPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] != "dhcp" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: dhcp/<pool_name>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}

func (r *dhcpPoolResource) modelToOptions(ctx context.Context, plan dhcpPoolModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.Interface.IsNull() {
		options["interface"] = plan.Interface.ValueString()
	}
	if !plan.Start.IsNull() {
		options["start"] = plan.Start.ValueInt64()
	}
	if !plan.Limit.IsNull() {
		options["limit"] = plan.Limit.ValueInt64()
	}
	if !plan.Leasetime.IsNull() {
		options["leasetime"] = plan.Leasetime.ValueString()
	}
	if !plan.DHCPv4.IsNull() {
		options["dhcpv4"] = plan.DHCPv4.ValueString()
	}
	if !plan.DHCPv6.IsNull() {
		options["dhcpv6"] = plan.DHCPv6.ValueString()
	}
	if !plan.RA.IsNull() {
		options["ra"] = plan.RA.ValueString()
	}
	if !plan.RAFlags.IsNull() {
		var flagsList []string
		diagnostics := plan.RAFlags.ElementsAs(ctx, &flagsList, false)
		if !diagnostics.HasError() {
			options["ra_flags"] = strings.Join(flagsList, " ")
		}
	}
	if !plan.Ignore.IsNull() {
		if plan.Ignore.ValueBool() {
			options["ignore"] = "1"
		} else {
			options["ignore"] = "0"
		}
	}

	return options
}

func (r *dhcpPoolResource) optionsToModel(data map[string]interface{}, state *dhcpPoolModel) {
	if v, ok := data["name"].(string); ok {
		state.Name = types.StringValue(v)
	} else if v, ok := data[".name"].(string); ok {
		state.Name = types.StringValue(v)
	}
	if v, ok := data["interface"].(string); ok {
		state.Interface = types.StringValue(v)
	}
	if v, ok := data["start"]; ok {
		if f, ok := v.(float64); ok {
			state.Start = types.Int64Value(int64(f))
		} else if s, ok := v.(string); ok {
			if i, err := strconv.Atoi(s); err == nil {
				state.Start = types.Int64Value(int64(i))
			}
		}
	} else {
		state.Start = types.Int64Value(100)
	}
	if v, ok := data["limit"]; ok {
		if f, ok := v.(float64); ok {
			state.Limit = types.Int64Value(int64(f))
		} else if s, ok := v.(string); ok {
			if i, err := strconv.Atoi(s); err == nil {
				state.Limit = types.Int64Value(int64(i))
			}
		}
	} else {
		state.Limit = types.Int64Value(150)
	}
	if v, ok := data["leasetime"].(string); ok {
		state.Leasetime = types.StringValue(v)
	} else {
		state.Leasetime = types.StringValue("12h")
	}
	if v, ok := data["dhcpv4"].(string); ok {
		state.DHCPv4 = types.StringValue(v)
	} else {
		state.DHCPv4 = types.StringValue("server")
	}
	if v, ok := data["dhcpv6"].(string); ok {
		state.DHCPv6 = types.StringValue(v)
	}
	if v, ok := data["ra"].(string); ok {
		state.RA = types.StringValue(v)
	}
	if flags, ok := data["ra_flags"]; ok {
		switch v := flags.(type) {
		case []interface{}:
			var flagsList []string
			for _, f := range v {
				if s, ok := f.(string); ok {
					flagsList = append(flagsList, s)
				}
			}
			state.RAFlags, _ = types.ListValueFrom(context.Background(), types.StringType, flagsList)
		case string:
			if v != "" {
				flagsList := strings.Split(v, " ")
				state.RAFlags, _ = types.ListValueFrom(context.Background(), types.StringType, flagsList)
			}
		}
	}
	if v, ok := data["ignore"].(string); ok {
		state.Ignore = types.BoolValue(v == "1" || v == "true")
	}
}
