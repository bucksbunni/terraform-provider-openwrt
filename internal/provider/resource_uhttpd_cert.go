package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewUHTTPdCertResource() resource.Resource {
	return &uhttpdCertResource{}
}

type uhttpdCertResource struct {
	client *JsonRpcClient
}

type uhttpdCertModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Days       types.Int64  `tfsdk:"days"`
	KeyType    types.String `tfsdk:"key_type"`
	Bits       types.Int64  `tfsdk:"bits"`
	ECCurve    types.String `tfsdk:"ec_curve"`
	Country    types.String `tfsdk:"country"`
	State      types.String `tfsdk:"state"`
	Location   types.String `tfsdk:"location"`
	CommonName types.String `tfsdk:"commonname"`
}

func (r *uhttpdCertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_uhttpd_cert"
}

func (r *uhttpdCertResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages OpenWrt uHTTPd certificate generation settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: uhttpd/<cert_name>.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Certificate instance name (e.g., 'defaults').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"days": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(397),
				Description: "Certificate validity in days.",
			},
			"key_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ec"),
				Description: "Key type: 'rsa' or 'ec'.",
			},
			"bits": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(2048),
				Description: "RSA key size in bits (for RSA keys).",
			},
			"ec_curve": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("P-256"),
				Description: "EC curve name (for EC keys).",
			},
			"country": schema.StringAttribute{
				Required:    true,
				Description: "Country code (e.g., 'US').",
			},
			"state": schema.StringAttribute{
				Required:    true,
				Description: "State or province name.",
			},
			"location": schema.StringAttribute{
				Required:    true,
				Description: "City or location.",
			},
			"commonname": schema.StringAttribute{
				Required:    true,
				Description: "Common name (typically hostname or domain).",
			},
		},
	}
}

func (r *uhttpdCertResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *uhttpdCertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan uhttpdCertModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(plan)

	_, err := r.client.UCISection(ctx, "uhttpd", "cert", name, options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating uhttpd cert", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "uhttpd"); err != nil {
		resp.Diagnostics.AddError("Error committing uhttpd config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	plan.ID = types.StringValue(fmt.Sprintf("uhttpd/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *uhttpdCertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state uhttpdCertModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	data, err := r.client.UCIGetAll(ctx, "uhttpd", name)
	if err != nil || len(data) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(data, &state)
	state.ID = types.StringValue(fmt.Sprintf("uhttpd/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *uhttpdCertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan uhttpdCertModel
	var state uhttpdCertModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(plan)

	if err := r.client.UCITSet(ctx, "uhttpd", name, options); err != nil {
		resp.Diagnostics.AddError("Error updating uhttpd cert", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "uhttpd"); err != nil {
		resp.Diagnostics.AddError("Error committing uhttpd config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *uhttpdCertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state uhttpdCertModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	if err := r.client.UCIDelete(ctx, "uhttpd", name); err != nil {
		resp.Diagnostics.AddError("Error deleting uhttpd cert", err.Error())
		return
	}
	if err := r.client.UCICommit(ctx, "uhttpd"); err != nil {
		resp.Diagnostics.AddError("Error committing uhttpd config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.State.RemoveResource(ctx)
}

func (r *uhttpdCertResource) modelToOptions(plan uhttpdCertModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.Days.IsNull() {
		options["days"] = plan.Days.ValueInt64()
	}
	if !plan.KeyType.IsNull() {
		options["key_type"] = plan.KeyType.ValueString()
	}
	if !plan.Bits.IsNull() {
		options["bits"] = plan.Bits.ValueInt64()
	}
	if !plan.ECCurve.IsNull() {
		options["ec_curve"] = plan.ECCurve.ValueString()
	}
	if !plan.Country.IsNull() {
		options["country"] = plan.Country.ValueString()
	}
	if !plan.State.IsNull() {
		options["state"] = plan.State.ValueString()
	}
	if !plan.Location.IsNull() {
		options["location"] = plan.Location.ValueString()
	}
	if !plan.CommonName.IsNull() {
		options["commonname"] = plan.CommonName.ValueString()
	}

	return options
}

func (r *uhttpdCertResource) optionsToModel(data map[string]interface{}, state *uhttpdCertModel) {
	if v, ok := data["days"]; ok {
		if f, ok := v.(float64); ok {
			state.Days = types.Int64Value(int64(f))
		}
	}
	if v, ok := data["key_type"].(string); ok {
		state.KeyType = types.StringValue(v)
	}
	if v, ok := data["bits"]; ok {
		if f, ok := v.(float64); ok {
			state.Bits = types.Int64Value(int64(f))
		}
	}
	if v, ok := data["ec_curve"].(string); ok {
		state.ECCurve = types.StringValue(v)
	}
	if v, ok := data["country"].(string); ok {
		state.Country = types.StringValue(v)
	}
	if v, ok := data["state"].(string); ok {
		state.State = types.StringValue(v)
	}
	if v, ok := data["location"].(string); ok {
		state.Location = types.StringValue(v)
	}
	if v, ok := data["commonname"].(string); ok {
		state.CommonName = types.StringValue(v)
	}
}
