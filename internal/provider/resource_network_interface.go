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

func NewNetworkInterfaceResource() resource.Resource {
	return &networkInterfaceResource{}
}

type networkInterfaceResource struct {
	client *JsonRpcClient
}

type networkInterfaceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Proto       types.String `tfsdk:"proto"`
	Device      types.String `tfsdk:"device"`
	IPAddr      types.List   `tfsdk:"ipaddr"`
	Gateway     types.String `tfsdk:"gateway"`
	DNS         types.List   `tfsdk:"dns"`
	Metric      types.Int64  `tfsdk:"metric"`
	Delegate    types.Bool   `tfsdk:"delegate"`
	IP6Addr     types.List   `tfsdk:"ip6addr"`
	IP6Prefix   types.String `tfsdk:"ip6prefix"`
	IP6Assign   types.String `tfsdk:"ip6assign"`
	IP6Gateway  types.String `tfsdk:"ip6gateway"`
	Auto        types.String `tfsdk:"auto"`
	IfType      types.String `tfsdk:"type"`
	BridgeEmpty types.Bool   `tfsdk:"bridge_empty"`
}

func (r *networkInterfaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_interface"
}

func (r *networkInterfaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an OpenWrt network interface.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: network/<interface_name>.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the interface (e.g., 'lan', 'wan', 'guest').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"proto": schema.StringAttribute{
				Required:    true,
				Description: "Protocol: 'static', 'dhcp', 'dhcpv6', 'pppoe', 'wireguard', 'qmi', etc.",
			},
			"device": schema.StringAttribute{
				Optional:    true,
				Description: "Network device (e.g., 'eth0', 'br-lan', 'wg0').",
			},
			"ipaddr": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "IPv4 address(es) with prefix (e.g., ['192.168.1.1/24', '10.0.0.1/24']).",
			},
			"gateway": schema.StringAttribute{
				Optional:    true,
				Description: "Default gateway IP address.",
			},
			"dns": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "DNS server(s) (e.g., ['8.8.8.8', '1.1.1.1']).",
			},
			"metric": schema.Int64Attribute{
				Optional:    true,
				Description: "Route metric for this interface.",
			},
			"delegate": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable IPv6 delegation (default: true).",
			},
			"ip6addr": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "IPv6 address(es) with prefix (e.g., ['fd00::1/64', '2001:db8::1/64']).",
			},
			"ip6prefix": schema.StringAttribute{
				Optional:    true,
				Description: "IPv6 prefix for downstream delegation.",
			},
			"ip6assign": schema.StringAttribute{
				Optional:    true,
				Description: "IPv6 prefix length to assign to downstream (e.g., '60').",
			},
			"ip6gateway": schema.StringAttribute{
				Optional:    true,
				Description: "IPv6 default gateway.",
			},
			"auto": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("1"),
				Description: "Enable interface on boot ('0' or '1').",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Description: "Interface type ('bridge', 'bonding', etc.).",
			},
			"bridge_empty": schema.BoolAttribute{
				Optional:    true,
				Description: "Allow empty bridge (no ports).",
			},
		},
	}
}

func (r *networkInterfaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *networkInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan networkInterfaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(ctx, plan)

	tflog.Debug(ctx, "Creating network interface", map[string]interface{}{"name": name})

	if err := r.client.UCISetSection(ctx, "network", name, "interface"); err != nil {
		resp.Diagnostics.AddError("Error creating network interface", err.Error())
		return
	}

	for key, value := range options {
		if err := r.client.UCISet(ctx, "network", name, key, value); err != nil {
			resp.Diagnostics.AddError("Error setting network interface option", err.Error())
			return
		}
	}

	if err := r.client.UCICommit(ctx, "network"); err != nil {
		resp.Diagnostics.AddError("Error committing network config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue(fmt.Sprintf("network/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *networkInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state networkInterfaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	ifaces, err := r.client.UCIForeach(ctx, "network", "interface")
	if err != nil {
		resp.Diagnostics.AddError("Error reading network interfaces", err.Error())
		return
	}

	var data map[string]interface{}
	for _, iface := range ifaces {
		if iface[".name"] == name {
			data = iface
			break
		}
	}

	if data == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(ctx, data, &state)
	state.ID = types.StringValue(fmt.Sprintf("network/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *networkInterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan networkInterfaceModel
	var state networkInterfaceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(ctx, plan)

	// A network interface is a named UCI section whose identifier is the
	// interface name itself, so address it directly.
	secName := name

	for key, value := range options {
		if err := r.client.UCISet(ctx, "network", secName, key, value); err != nil {
			resp.Diagnostics.AddError("Error updating network interface option", err.Error())
			return
		}
	}

	if err := r.client.UCICommit(ctx, "network"); err != nil {
		resp.Diagnostics.AddError("Error committing network config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue(fmt.Sprintf("network/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *networkInterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state networkInterfaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	if err := r.client.UCIDelete(ctx, "network", name); err != nil {
		resp.Diagnostics.AddError("Error deleting network interface", err.Error())
		return
	}
	if err := r.client.UCICommit(ctx, "network"); err != nil {
		resp.Diagnostics.AddError("Error committing network config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.State.RemoveResource(ctx)
}

func (r *networkInterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] != "network" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: network/<interface_name>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}

func (r *networkInterfaceResource) modelToOptions(ctx context.Context, plan networkInterfaceModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.Proto.IsNull() {
		options["proto"] = plan.Proto.ValueString()
	}
	if !plan.Device.IsNull() {
		options["device"] = plan.Device.ValueString()
	}
	if !plan.IPAddr.IsNull() {
		var addrs []string
		plan.IPAddr.ElementsAs(ctx, &addrs, false)
		// Convert to []interface{} for UCI list directive
		list := make([]interface{}, len(addrs))
		for i, a := range addrs {
			list[i] = a
		}
		options["ipaddr"] = list
	}
	if !plan.Gateway.IsNull() {
		options["gateway"] = plan.Gateway.ValueString()
	}
	if !plan.DNS.IsNull() {
		var dnsList []string
		plan.DNS.ElementsAs(ctx, &dnsList, false)
		options["dns"] = strings.Join(dnsList, " ")
	}
	if !plan.Metric.IsNull() {
		options["metric"] = plan.Metric.ValueInt64()
	}
	if !plan.Delegate.IsNull() {
		if plan.Delegate.ValueBool() {
			options["delegate"] = "1"
		} else {
			options["delegate"] = "0"
		}
	}
	if !plan.IP6Addr.IsNull() {
		var addrs []string
		plan.IP6Addr.ElementsAs(ctx, &addrs, false)
		// Convert to []interface{} for UCI list directive
		list := make([]interface{}, len(addrs))
		for i, a := range addrs {
			list[i] = a
		}
		options["ip6addr"] = list
	}
	if !plan.IP6Prefix.IsNull() {
		options["ip6prefix"] = plan.IP6Prefix.ValueString()
	}
	if !plan.IP6Assign.IsNull() {
		options["ip6assign"] = plan.IP6Assign.ValueString()
	}
	if !plan.IP6Gateway.IsNull() {
		options["ip6gateway"] = plan.IP6Gateway.ValueString()
	}
	if !plan.Auto.IsNull() {
		options["auto"] = plan.Auto.ValueString()
	}
	if !plan.IfType.IsNull() {
		options["type"] = plan.IfType.ValueString()
	}
	if !plan.BridgeEmpty.IsNull() {
		if plan.BridgeEmpty.ValueBool() {
			options["bridge_empty"] = "1"
		} else {
			options["bridge_empty"] = "0"
		}
	}

	return options
}

func (r *networkInterfaceResource) optionsToModel(ctx context.Context, data map[string]interface{}, state *networkInterfaceModel) {
	if v, ok := data["proto"].(string); ok {
		state.Proto = types.StringValue(v)
	}
	if v, ok := data["device"].(string); ok {
		state.Device = types.StringValue(v)
	}
	if v, ok := data["ipaddr"].([]interface{}); ok {
		var addrs []string
		for _, a := range v {
			if s, ok := a.(string); ok {
				addrs = append(addrs, s)
			}
		}
		if len(addrs) > 0 {
			state.IPAddr, _ = types.ListValueFrom(ctx, types.StringType, addrs)
		}
	} else if v, ok := data["ipaddr"].(string); ok && v != "" {
		addrs := strings.Split(v, " ")
		state.IPAddr, _ = types.ListValueFrom(ctx, types.StringType, addrs)
	}
	if v, ok := data["gateway"].(string); ok {
		state.Gateway = types.StringValue(v)
	}
	if v, ok := data["dns"].(string); ok && v != "" {
		dnsList := strings.Split(v, " ")
		state.DNS, _ = types.ListValueFrom(ctx, types.StringType, dnsList)
	}
	if v, ok := data["metric"]; ok {
		if f, ok := v.(float64); ok {
			state.Metric = types.Int64Value(int64(f))
		}
	}
	if v, ok := data["delegate"].(string); ok {
		state.Delegate = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["ip6addr"].([]interface{}); ok {
		var addrs []string
		for _, a := range v {
			if s, ok := a.(string); ok {
				addrs = append(addrs, s)
			}
		}
		if len(addrs) > 0 {
			state.IP6Addr, _ = types.ListValueFrom(ctx, types.StringType, addrs)
		}
	} else if v, ok := data["ip6addr"].(string); ok && v != "" {
		addrs := strings.Split(v, " ")
		state.IP6Addr, _ = types.ListValueFrom(ctx, types.StringType, addrs)
	}
	if v, ok := data["ip6prefix"].(string); ok {
		state.IP6Prefix = types.StringValue(v)
	}
	if v, ok := data["ip6assign"].(string); ok {
		state.IP6Assign = types.StringValue(v)
	}
	if v, ok := data["ip6gateway"].(string); ok {
		state.IP6Gateway = types.StringValue(v)
	}
	if v, ok := data["auto"].(string); ok {
		state.Auto = types.StringValue(v)
	}
	if v, ok := data["type"].(string); ok {
		state.IfType = types.StringValue(v)
	}
	if v, ok := data["bridge_empty"].(string); ok {
		state.BridgeEmpty = types.BoolValue(v == "1" || v == "true")
	}
}
