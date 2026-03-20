package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewDHCPOdhcpdResource() resource.Resource {
	return &dhcpOdhcpdResource{}
}

type dhcpOdhcpdResource struct {
	client *JsonRpcClient
}

type dhcpOdhcpdModel struct {
	ID           types.String `tfsdk:"id"`
	MainDHCP     types.Bool   `tfsdk:"maindhcp"`
	LeaseFile    types.String `tfsdk:"leasefile"`
	LeaseTrigger types.String `tfsdk:"leasetrigger"`
	LogLevel     types.Int64  `tfsdk:"loglevel"`
}

func (r *dhcpOdhcpdResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_odhcpd"
}

func (r *dhcpOdhcpdResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages OpenWrt DHCPv6 (odhcpd) settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: dhcp/odhcpd.",
			},
			"maindhcp": schema.BoolAttribute{
				Optional:    true,
				Description: "Run odhcpd as DHCPv6 server on interface without RA.",
			},
			"leasefile": schema.StringAttribute{
				Optional:    true,
				Description: "Lease file path for DHCPv6.",
			},
			"leasetrigger": schema.StringAttribute{
				Optional:    true,
				Description: "Script to run on lease changes.",
			},
			"loglevel": schema.Int64Attribute{
				Optional:    true,
				Description: "Log level (0-8).",
			},
		},
	}
}

func (r *dhcpOdhcpdResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dhcpOdhcpdResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dhcpOdhcpdModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := r.modelToOptions(plan)

	_, err := r.client.UCISection(ctx, "dhcp", "odhcpd", "odhcpd", options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating odhcpd config", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "dhcp"); err != nil {
		resp.Diagnostics.AddError("Error committing DHCP config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue("dhcp/odhcpd")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dhcpOdhcpdResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dhcpOdhcpdModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data, err := r.client.UCIGetAll(ctx, "dhcp", "odhcpd")
	if err != nil || len(data) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(data, &state)
	state.ID = types.StringValue("dhcp/odhcpd")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dhcpOdhcpdResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dhcpOdhcpdModel
	var state dhcpOdhcpdModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := r.modelToOptions(plan)

	if err := r.client.UCITSet(ctx, "dhcp", "odhcpd", options); err != nil {
		resp.Diagnostics.AddError("Error updating odhcpd config", err.Error())
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

func (r *dhcpOdhcpdResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Delete not supported", "odhcpd settings cannot be deleted, only modified.")
}

func (r *dhcpOdhcpdResource) modelToOptions(plan dhcpOdhcpdModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.MainDHCP.IsNull() {
		options["maindhcp"] = boolToString(plan.MainDHCP.ValueBool())
	}
	if !plan.LeaseFile.IsNull() {
		options["leasefile"] = plan.LeaseFile.ValueString()
	}
	if !plan.LeaseTrigger.IsNull() {
		options["leasetrigger"] = plan.LeaseTrigger.ValueString()
	}
	if !plan.LogLevel.IsNull() {
		options["loglevel"] = plan.LogLevel.ValueInt64()
	}

	return options
}

func (r *dhcpOdhcpdResource) optionsToModel(data map[string]interface{}, state *dhcpOdhcpdModel) {
	if v, ok := data["maindhcp"].(string); ok {
		state.MainDHCP = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["leasefile"].(string); ok {
		state.LeaseFile = types.StringValue(v)
	}
	if v, ok := data["leasetrigger"].(string); ok {
		state.LeaseTrigger = types.StringValue(v)
	}
	if v, ok := data["loglevel"]; ok {
		if f, ok := v.(float64); ok {
			state.LogLevel = types.Int64Value(int64(f))
		}
	}
}
