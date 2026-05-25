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

func NewFirewallRuleResource() resource.Resource {
	return &firewallRuleResource{}
}

type firewallRuleResource struct {
	client *JsonRpcClient
}

type firewallRuleModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Src      types.String `tfsdk:"src"`
	Dest     types.String `tfsdk:"dest"`
	Proto    types.String `tfsdk:"proto"`
	SrcPort  types.String `tfsdk:"src_port"`
	DestPort types.String `tfsdk:"dest_port"`
	SrcIP    types.String `tfsdk:"src_ip"`
	DestIP   types.String `tfsdk:"dest_ip"`
	Target   types.String `tfsdk:"target"`
	Family   types.String `tfsdk:"family"`
	ICMPType types.List   `tfsdk:"icmp_type"`
	Limit    types.String `tfsdk:"limit"`
	Extra    types.String `tfsdk:"extra"`
	Enabled  types.Bool   `tfsdk:"enabled"`
}

func (r *firewallRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_rule"
}

func (r *firewallRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an OpenWrt firewall rule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: firewall/<rule_name>.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the firewall rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"src": schema.StringAttribute{
				Optional:    true,
				Description: "Source zone name.",
			},
			"dest": schema.StringAttribute{
				Optional:    true,
				Description: "Destination zone name.",
			},
			"proto": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("tcpudp"),
				Description: "Protocol: 'tcp', 'udp', 'tcpudp', 'icmp', 'all'.",
			},
			"src_port": schema.StringAttribute{
				Optional:    true,
				Description: "Source port or range (e.g., '1024:65535').",
			},
			"dest_port": schema.StringAttribute{
				Optional:    true,
				Description: "Destination port or range (e.g., '53', '80:443').",
			},
			"src_ip": schema.StringAttribute{
				Optional:    true,
				Description: "Source IP address or CIDR.",
			},
			"dest_ip": schema.StringAttribute{
				Optional:    true,
				Description: "Destination IP address or CIDR.",
			},
			"target": schema.StringAttribute{
				Required:    true,
				Description: "Target action: 'ACCEPT', 'REJECT', 'DROP'.",
			},
			"family": schema.StringAttribute{
				Optional:    true,
				Description: "IP family: 'ipv4', 'ipv6', or 'all'.",
			},
			"icmp_type": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "ICMP type (e.g., ['echo-request', 'echo-reply']).",
			},
			"limit": schema.StringAttribute{
				Optional:    true,
				Description: "Rate limit (e.g., '1000/sec').",
			},
			"extra": schema.StringAttribute{
				Optional:    true,
				Description: "Extra iptables options.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable or disable this rule.",
			},
		},
	}
}

func (r *firewallRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *firewallRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan firewallRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(ctx, plan)

	secName, err := r.client.UCISection(ctx, "firewall", "rule", name, options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating firewall rule", err.Error())
		return
	}

	if name == "" {
		name = secName
	}

	if err := r.client.UCICommit(ctx, "firewall"); err != nil {
		resp.Diagnostics.AddError("Error committing firewall config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue(fmt.Sprintf("firewall/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *firewallRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state firewallRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	data, err := r.client.UCIGetAll(ctx, "firewall", name)
	if err != nil || len(data) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(data, &state)
	state.ID = types.StringValue(fmt.Sprintf("firewall/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan firewallRuleModel
	var state firewallRuleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(ctx, plan)

	if err := r.client.UCITSet(ctx, "firewall", name, options); err != nil {
		resp.Diagnostics.AddError("Error updating firewall rule", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "firewall"); err != nil {
		resp.Diagnostics.AddError("Error committing firewall config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *firewallRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state firewallRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	if err := r.client.UCIDelete(ctx, "firewall", name); err != nil {
		resp.Diagnostics.AddError("Error deleting firewall rule", err.Error())
		return
	}
	if err := r.client.UCICommit(ctx, "firewall"); err != nil {
		resp.Diagnostics.AddError("Error committing firewall config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.State.RemoveResource(ctx)
}

func (r *firewallRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] != "firewall" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: firewall/<rule_name>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}

func (r *firewallRuleResource) modelToOptions(ctx context.Context, plan firewallRuleModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.Src.IsNull() {
		options["src"] = plan.Src.ValueString()
	}
	if !plan.Dest.IsNull() {
		options["dest"] = plan.Dest.ValueString()
	}
	if !plan.Proto.IsNull() {
		options["proto"] = plan.Proto.ValueString()
	}
	if !plan.SrcPort.IsNull() {
		options["src_port"] = plan.SrcPort.ValueString()
	}
	if !plan.DestPort.IsNull() {
		options["dest_port"] = plan.DestPort.ValueString()
	}
	if !plan.SrcIP.IsNull() {
		options["src_ip"] = plan.SrcIP.ValueString()
	}
	if !plan.DestIP.IsNull() {
		options["dest_ip"] = plan.DestIP.ValueString()
	}
	if !plan.Target.IsNull() {
		options["target"] = plan.Target.ValueString()
	}
	if !plan.Family.IsNull() {
		options["family"] = plan.Family.ValueString()
	}
	if !plan.ICMPType.IsNull() {
		var icmpList []string
		diagnostics := plan.ICMPType.ElementsAs(ctx, &icmpList, false)
		if !diagnostics.HasError() {
			options["icmp_type"] = strings.Join(icmpList, " ")
		}
	}
	if !plan.Limit.IsNull() {
		options["limit"] = plan.Limit.ValueString()
	}
	if !plan.Extra.IsNull() {
		options["extra"] = plan.Extra.ValueString()
	}
	if !plan.Enabled.IsNull() {
		if plan.Enabled.ValueBool() {
			options["enabled"] = "1"
		} else {
			options["enabled"] = "0"
		}
	}

	return options
}

func (r *firewallRuleResource) optionsToModel(data map[string]interface{}, state *firewallRuleModel) {
	if v, ok := data["name"].(string); ok {
		state.Name = types.StringValue(v)
	}
	if v, ok := data["src"].(string); ok {
		state.Src = types.StringValue(v)
	}
	if v, ok := data["dest"].(string); ok {
		state.Dest = types.StringValue(v)
	}
	if v, ok := data["proto"].(string); ok {
		state.Proto = types.StringValue(v)
	}
	if v, ok := data["src_port"].(string); ok {
		state.SrcPort = types.StringValue(v)
	}
	if v, ok := data["dest_port"].(string); ok {
		state.DestPort = types.StringValue(v)
	}
	if v, ok := data["src_ip"].(string); ok {
		state.SrcIP = types.StringValue(v)
	}
	if v, ok := data["dest_ip"].(string); ok {
		state.DestIP = types.StringValue(v)
	}
	if v, ok := data["target"].(string); ok {
		state.Target = types.StringValue(v)
	}
	if v, ok := data["family"].(string); ok {
		state.Family = types.StringValue(v)
	}
	if icmp, ok := data["icmp_type"]; ok {
		switch v := icmp.(type) {
		case []interface{}:
			var icmpList []string
			for _, i := range v {
				if s, ok := i.(string); ok {
					icmpList = append(icmpList, s)
				}
			}
			state.ICMPType, _ = types.ListValueFrom(context.Background(), types.StringType, icmpList)
		case string:
			if v != "" {
				icmpList := strings.Split(v, " ")
				state.ICMPType, _ = types.ListValueFrom(context.Background(), types.StringType, icmpList)
			}
		}
	}
	if v, ok := data["limit"].(string); ok {
		state.Limit = types.StringValue(v)
	}
	if v, ok := data["extra"].(string); ok {
		state.Extra = types.StringValue(v)
	}
	if v, ok := data["enabled"].(string); ok {
		state.Enabled = types.BoolValue(v == "1" || v == "true")
	}
}
