package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewUHTTPdResource() resource.Resource {
	return &uhttpdResource{}
}

type uhttpdResource struct {
	client *JsonRpcClient
}

type uhttpdModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	ListenHTTP     types.String `tfsdk:"listen_http"`
	ListenHTTPS    types.String `tfsdk:"listen_https"`
	RedirectHTTPS  types.Bool   `tfsdk:"redirect_https"`
	Home           types.String `tfsdk:"home"`
	RFC1918Filter  types.Bool   `tfsdk:"rfc1918_filter"`
	MaxRequests    types.Int64  `tfsdk:"max_requests"`
	MaxConnections types.Int64  `tfsdk:"max_connections"`
	Cert           types.String `tfsdk:"cert"`
	Key            types.String `tfsdk:"key"`
	CGIPrefix      types.String `tfsdk:"cgi_prefix"`
	LuaPrefix      types.String `tfsdk:"lua_prefix"`
	ScriptTimeout  types.Int64  `tfsdk:"script_timeout"`
	NetworkTimeout types.Int64  `tfsdk:"network_timeout"`
	HTTPKeepAlive  types.Int64  `tfsdk:"http_keepalive"`
	TCPKeepAlive   types.Int64  `tfsdk:"tcp_keepalive"`
	UbusPrefix     types.String `tfsdk:"ubus_prefix"`
}

func (r *uhttpdResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_uhttpd"
}

func (r *uhttpdResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages OpenWrt uHTTPd web server settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: uhttpd/main.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Instance name (e.g., 'main').",
			},
			"listen_http": schema.StringAttribute{
				Optional:    true,
				Description: "HTTP listen addresses, space-separated (e.g., '0.0.0.0:80 [::]:80').",
			},
			"listen_https": schema.StringAttribute{
				Optional:    true,
				Description: "HTTPS listen addresses, space-separated.",
			},
			"redirect_https": schema.BoolAttribute{
				Optional:    true,
				Description: "Redirect HTTP to HTTPS.",
			},
			"home": schema.StringAttribute{
				Optional:    true,
				Description: "Document root directory.",
			},
			"rfc1918_filter": schema.BoolAttribute{
				Optional:    true,
				Description: "Filter requests to private IP ranges.",
			},
			"max_requests": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum concurrent requests.",
			},
			"max_connections": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum connections.",
			},
			"cert": schema.StringAttribute{
				Optional:    true,
				Description: "HTTPS certificate path.",
			},
			"key": schema.StringAttribute{
				Optional:    true,
				Description: "HTTPS private key path.",
			},
			"cgi_prefix": schema.StringAttribute{
				Optional:    true,
				Description: "CGI handler prefix.",
			},
			"lua_prefix": schema.StringAttribute{
				Optional:    true,
				Description: "Lua CGI handler prefix.",
			},
			"script_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "CGI script timeout in seconds.",
			},
			"network_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "Network timeout in seconds.",
			},
			"http_keepalive": schema.Int64Attribute{
				Optional:    true,
				Description: "HTTP keepalive timeout.",
			},
			"tcp_keepalive": schema.Int64Attribute{
				Optional:    true,
				Description: "TCP keepalive.",
			},
			"ubus_prefix": schema.StringAttribute{
				Optional:    true,
				Description: "Ubus API prefix.",
			},
		},
	}
}

func (r *uhttpdResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *uhttpdResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan uhttpdModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(plan)

	_, err := r.client.UCISection(ctx, "uhttpd", "uhttpd", name, options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating uhttpd config", err.Error())
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

func (r *uhttpdResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state uhttpdModel
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

func (r *uhttpdResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan uhttpdModel
	var state uhttpdModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(plan)

	if err := r.client.UCITSet(ctx, "uhttpd", name, options); err != nil {
		resp.Diagnostics.AddError("Error updating uhttpd config", err.Error())
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

func (r *uhttpdResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state uhttpdModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	if err := r.client.UCIDelete(ctx, "uhttpd", name); err != nil {
		resp.Diagnostics.AddError("Error deleting uhttpd config", err.Error())
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

func (r *uhttpdResource) modelToOptions(plan uhttpdModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.ListenHTTP.IsNull() {
		options["listen_http"] = plan.ListenHTTP.ValueString()
	}
	if !plan.ListenHTTPS.IsNull() {
		options["listen_https"] = plan.ListenHTTPS.ValueString()
	}
	if !plan.RedirectHTTPS.IsNull() {
		options["redirect_https"] = boolToString(plan.RedirectHTTPS.ValueBool())
	}
	if !plan.Home.IsNull() {
		options["home"] = plan.Home.ValueString()
	}
	if !plan.RFC1918Filter.IsNull() {
		options["rfc1918_filter"] = boolToString(plan.RFC1918Filter.ValueBool())
	}
	if !plan.MaxRequests.IsNull() {
		options["max_requests"] = plan.MaxRequests.ValueInt64()
	}
	if !plan.MaxConnections.IsNull() {
		options["max_connections"] = plan.MaxConnections.ValueInt64()
	}
	if !plan.Cert.IsNull() {
		options["cert"] = plan.Cert.ValueString()
	}
	if !plan.Key.IsNull() {
		options["key"] = plan.Key.ValueString()
	}
	if !plan.CGIPrefix.IsNull() {
		options["cgi_prefix"] = plan.CGIPrefix.ValueString()
	}
	if !plan.LuaPrefix.IsNull() {
		options["lua_prefix"] = plan.LuaPrefix.ValueString()
	}
	if !plan.ScriptTimeout.IsNull() {
		options["script_timeout"] = plan.ScriptTimeout.ValueInt64()
	}
	if !plan.NetworkTimeout.IsNull() {
		options["network_timeout"] = plan.NetworkTimeout.ValueInt64()
	}
	if !plan.HTTPKeepAlive.IsNull() {
		options["http_keepalive"] = plan.HTTPKeepAlive.ValueInt64()
	}
	if !plan.TCPKeepAlive.IsNull() {
		options["tcp_keepalive"] = plan.TCPKeepAlive.ValueInt64()
	}
	if !plan.UbusPrefix.IsNull() {
		options["ubus_prefix"] = plan.UbusPrefix.ValueString()
	}

	return options
}

func (r *uhttpdResource) optionsToModel(data map[string]interface{}, state *uhttpdModel) {
	if v, ok := data["listen_http"].(string); ok {
		state.ListenHTTP = types.StringValue(v)
	}
	if v, ok := data["listen_https"].(string); ok {
		state.ListenHTTPS = types.StringValue(v)
	}
	if v, ok := data["redirect_https"].(string); ok {
		state.RedirectHTTPS = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["home"].(string); ok {
		state.Home = types.StringValue(v)
	}
	if v, ok := data["rfc1918_filter"].(string); ok {
		state.RFC1918Filter = types.BoolValue(v == "1" || v == "true")
	}
	if v, ok := data["max_requests"]; ok {
		if f, ok := v.(float64); ok {
			state.MaxRequests = types.Int64Value(int64(f))
		}
	}
	if v, ok := data["max_connections"]; ok {
		if f, ok := v.(float64); ok {
			state.MaxConnections = types.Int64Value(int64(f))
		}
	}
	if v, ok := data["cert"].(string); ok {
		state.Cert = types.StringValue(v)
	}
	if v, ok := data["key"].(string); ok {
		state.Key = types.StringValue(v)
	}
	if v, ok := data["cgi_prefix"].(string); ok {
		state.CGIPrefix = types.StringValue(v)
	}
	if v, ok := data["lua_prefix"].(string); ok {
		state.LuaPrefix = types.StringValue(v)
	}
	if v, ok := data["script_timeout"]; ok {
		if f, ok := v.(float64); ok {
			state.ScriptTimeout = types.Int64Value(int64(f))
		}
	}
	if v, ok := data["network_timeout"]; ok {
		if f, ok := v.(float64); ok {
			state.NetworkTimeout = types.Int64Value(int64(f))
		}
	}
	if v, ok := data["http_keepalive"]; ok {
		if f, ok := v.(float64); ok {
			state.HTTPKeepAlive = types.Int64Value(int64(f))
		}
	}
	if v, ok := data["tcp_keepalive"]; ok {
		if f, ok := v.(float64); ok {
			state.TCPKeepAlive = types.Int64Value(int64(f))
		}
	}
	if v, ok := data["ubus_prefix"].(string); ok {
		state.UbusPrefix = types.StringValue(v)
	}
}
