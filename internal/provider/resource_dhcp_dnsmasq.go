package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewDHCPDNSMasqResource() resource.Resource {
	return &dhcpDNSMasqResource{}
}

type dhcpDNSMasqResource struct {
	client *JsonRpcClient
}

type dhcpDNSMasqModel struct {
	ID               types.String `tfsdk:"id"`
	DomainNeeded     types.Bool   `tfsdk:"domainneeded"`
	LocaliseQueries  types.Bool   `tfsdk:"localise_queries"`
	RebindProtection types.Bool   `tfsdk:"rebind_protection"`
	RebindLocalhost  types.Bool   `tfsdk:"rebind_localhost"`
	Local            types.String `tfsdk:"local"`
	Domain           types.String `tfsdk:"domain"`
	ExpandHosts      types.Bool   `tfsdk:"expand_hosts"`
	CacheSize        types.Int64  `tfsdk:"cachesize"`
	Authoritative    types.Bool   `tfsdk:"authoritative"`
	ReadEthers       types.Bool   `tfsdk:"readethers"`
	LeaseFile        types.String `tfsdk:"leasefile"`
	ResolvFile       types.String `tfsdk:"resolvfile"`
	LocalService     types.Bool   `tfsdk:"localservice"`
	EDNSPacketMax    types.Int64  `tfsdk:"ednspacket_max"`
	ConfDir          types.String `tfsdk:"confdir"`
	RebindDomain     types.List   `tfsdk:"rebind_domain"`
	Server           types.List   `tfsdk:"server"`
}

func (r *dhcpDNSMasqResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_dnsmasq"
}

func (r *dhcpDNSMasqResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages DNS/DHCP settings (dnsmasq) for OpenWrt.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: dhcp/dnsmasq.",
			},
			"domainneeded": schema.BoolAttribute{
				Optional:    true,
				Description: "Require domain names to be valid.",
			},
			"localise_queries": schema.BoolAttribute{
				Optional:    true,
				Description: "Answer DNS queries based on the subnet of the requester.",
			},
			"rebind_protection": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable DNS rebind attack protection.",
			},
			"rebind_localhost": schema.BoolAttribute{
				Optional:    true,
				Description: "Allow DNS rebind for 127.0.0.1.",
			},
			"local": schema.StringAttribute{
				Optional:    true,
				Description: "Local domain for DNS queries (e.g., '/lan/').",
			},
			"domain": schema.StringAttribute{
				Optional:    true,
				Description: "Local domain name.",
			},
			"expand_hosts": schema.BoolAttribute{
				Optional:    true,
				Description: "Add local domain suffix to hostnames.",
			},
			"cachesize": schema.Int64Attribute{
				Optional:    true,
				Description: "DNS cache size.",
			},
			"authoritative": schema.BoolAttribute{
				Optional:    true,
				Description: "Only respond to DHCP requests for this device.",
			},
			"readethers": schema.BoolAttribute{
				Optional:    true,
				Description: "Read /etc/ethers for static leases.",
			},
			"leasefile": schema.StringAttribute{
				Optional:    true,
				Description: "DHCP lease file path.",
			},
			"resolvfile": schema.StringAttribute{
				Optional:    true,
				Description: "DNS resolver configuration file path.",
			},
			"localservice": schema.BoolAttribute{
				Optional:    true,
				Description: "Only serve DNS to local subnets.",
			},
			"ednspacket_max": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum EDNS.0 packet size.",
			},
			"confdir": schema.StringAttribute{
				Optional:    true,
				Description: "Additional configuration directory.",
			},
			"rebind_domain": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Domain whitelist for rebind protection (e.g., ['example.com']).",
			},
			"server": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Upstream DNS servers (e.g., ['8.8.8.8', '1.1.1.1']).",
			},
		},
	}
}

func (r *dhcpDNSMasqResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dhcpDNSMasqResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dhcpDNSMasqModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := r.modelToOptions(ctx, plan)

	_, err := r.client.UCISection(ctx, "dhcp", "dnsmasq", "dnsmasq", options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating dnsmasq config", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "dhcp"); err != nil {
		resp.Diagnostics.AddError("Error committing DHCP config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue("dhcp/dnsmasq")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dhcpDNSMasqResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dhcpDNSMasqModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data, err := r.client.UCIGetAll(ctx, "dhcp", "dnsmasq")
	if err != nil || len(data) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(data, &state)
	state.ID = types.StringValue("dhcp/dnsmasq")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dhcpDNSMasqResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dhcpDNSMasqModel
	var state dhcpDNSMasqModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := r.modelToOptions(ctx, plan)

	if err := r.client.UCITSet(ctx, "dhcp", "dnsmasq", options); err != nil {
		resp.Diagnostics.AddError("Error updating dnsmasq config", err.Error())
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

func (r *dhcpDNSMasqResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dhcpDNSMasqModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UCIDelete(ctx, "dhcp", "dnsmasq"); err != nil {
		resp.Diagnostics.AddError("Error deleting dnsmasq config", err.Error())
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

func (r *dhcpDNSMasqResource) modelToOptions(ctx context.Context, plan dhcpDNSMasqModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.DomainNeeded.IsNull() {
		options["domainneeded"] = boolToString(plan.DomainNeeded.ValueBool())
	}
	if !plan.LocaliseQueries.IsNull() {
		options["localise_queries"] = boolToString(plan.LocaliseQueries.ValueBool())
	}
	if !plan.RebindProtection.IsNull() {
		options["rebind_protection"] = boolToString(plan.RebindProtection.ValueBool())
	}
	if !plan.RebindLocalhost.IsNull() {
		options["rebind_localhost"] = boolToString(plan.RebindLocalhost.ValueBool())
	}
	if !plan.Local.IsNull() {
		options["local"] = plan.Local.ValueString()
	}
	if !plan.Domain.IsNull() {
		options["domain"] = plan.Domain.ValueString()
	}
	if !plan.ExpandHosts.IsNull() {
		options["expandhosts"] = boolToString(plan.ExpandHosts.ValueBool())
	}
	if !plan.CacheSize.IsNull() {
		options["cachesize"] = plan.CacheSize.ValueInt64()
	}
	if !plan.Authoritative.IsNull() {
		options["authoritative"] = boolToString(plan.Authoritative.ValueBool())
	}
	if !plan.ReadEthers.IsNull() {
		options["readethers"] = boolToString(plan.ReadEthers.ValueBool())
	}
	if !plan.LeaseFile.IsNull() {
		options["leasefile"] = plan.LeaseFile.ValueString()
	}
	if !plan.ResolvFile.IsNull() {
		options["resolvfile"] = plan.ResolvFile.ValueString()
	}
	if !plan.LocalService.IsNull() {
		options["localservice"] = boolToString(plan.LocalService.ValueBool())
	}
	if !plan.EDNSPacketMax.IsNull() {
		options["ednspacket_max"] = plan.EDNSPacketMax.ValueInt64()
	}
	if !plan.ConfDir.IsNull() {
		options["confdir"] = plan.ConfDir.ValueString()
	}
	if !plan.RebindDomain.IsNull() {
		var domainList []string
		diagnostics := plan.RebindDomain.ElementsAs(ctx, &domainList, false)
		if !diagnostics.HasError() {
			options["rebind_domain"] = strings.Join(domainList, " ")
		}
	}
	if !plan.Server.IsNull() {
		var serverList []string
		diagnostics := plan.Server.ElementsAs(ctx, &serverList, false)
		if !diagnostics.HasError() {
			options["server"] = strings.Join(serverList, " ")
		}
	}

	return options
}

func (r *dhcpDNSMasqResource) optionsToModel(data map[string]interface{}, state *dhcpDNSMasqModel) {
	if v, ok := data["domainneeded"].(string); ok {
		state.DomainNeeded = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["localise_queries"].(string); ok {
		state.LocaliseQueries = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["rebind_protection"].(string); ok {
		state.RebindProtection = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["rebind_localhost"].(string); ok {
		state.RebindLocalhost = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["local"].(string); ok {
		state.Local = types.StringValue(v)
	}
	if v, ok := data["domain"].(string); ok {
		state.Domain = types.StringValue(v)
	}
	if v, ok := data["expandhosts"].(string); ok {
		state.ExpandHosts = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["cachesize"]; ok {
		if f, ok := v.(float64); ok {
			state.CacheSize = types.Int64Value(int64(f))
		}
	}
	if v, ok := data["authoritative"].(string); ok {
		state.Authoritative = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["readethers"].(string); ok {
		state.ReadEthers = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["leasefile"].(string); ok {
		state.LeaseFile = types.StringValue(v)
	}
	if v, ok := data["resolvfile"].(string); ok {
		state.ResolvFile = types.StringValue(v)
	}
	if v, ok := data["localservice"].(string); ok {
		state.LocalService = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["ednspacket_max"]; ok {
		if f, ok := v.(float64); ok {
			state.EDNSPacketMax = types.Int64Value(int64(f))
		}
	}
	if v, ok := data["confdir"].(string); ok {
		state.ConfDir = types.StringValue(v)
	}
	if domain, ok := data["rebind_domain"]; ok {
		switch v := domain.(type) {
		case []interface{}:
			var domainList []string
			for _, d := range v {
				if s, ok := d.(string); ok {
					domainList = append(domainList, s)
				}
			}
			state.RebindDomain, _ = types.ListValueFrom(context.Background(), types.StringType, domainList)
		case string:
			if v != "" {
				domainList := strings.Split(v, " ")
				state.RebindDomain, _ = types.ListValueFrom(context.Background(), types.StringType, domainList)
			}
		}
	}
	if server, ok := data["server"]; ok {
		switch v := server.(type) {
		case []interface{}:
			var serverList []string
			for _, s := range v {
				if str, ok := s.(string); ok {
					serverList = append(serverList, str)
				}
			}
			state.Server, _ = types.ListValueFrom(context.Background(), types.StringType, serverList)
		case string:
			if v != "" {
				serverList := strings.Split(v, " ")
				state.Server, _ = types.ListValueFrom(context.Background(), types.StringType, serverList)
			}
		}
	}
}

func boolToString(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
