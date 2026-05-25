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

func NewNetworkWireguardResource() resource.Resource {
	return &networkWireguardResource{}
}

type networkWireguardResource struct {
	client *JsonRpcClient
}

type networkWireguardModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	PublicKey           types.String `tfsdk:"public_key"`
	EndpointHost        types.String `tfsdk:"endpoint_host"`
	EndpointPort        types.Int64  `tfsdk:"endpoint_port"`
	PersistentKeepalive types.Int64  `tfsdk:"persistent_keepalive"`
	AllowedIPs          types.List   `tfsdk:"allowed_ips"`
}

func (r *networkWireguardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_wireguard"
}

func (r *networkWireguardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a WireGuard peer in OpenWrt.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal ID: network/<peer_name>.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the WireGuard peer (e.g., 'wgclient1').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description or label for this peer.",
			},
			"public_key": schema.StringAttribute{
				Required:    true,
				Description: "WireGuard public key of the peer.",
			},
			"endpoint_host": schema.StringAttribute{
				Optional:    true,
				Description: "Endpoint hostname or IP address.",
			},
			"endpoint_port": schema.Int64Attribute{
				Optional:    true,
				Description: "Endpoint UDP port (default: 51820).",
			},
			"persistent_keepalive": schema.Int64Attribute{
				Optional:    true,
				Description: "Persistent keepalive interval in seconds (0 to disable).",
			},
			"allowed_ips": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Allowed IPs for this peer (e.g., ['10.0.0.2/32', '192.168.100.0/24']).",
			},
		},
	}
}

func (r *networkWireguardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *networkWireguardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan networkWireguardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(ctx, plan)

	peerType := fmt.Sprintf("wireguard_%s", extractInterfaceName(name))

	_, err := r.client.UCISection(ctx, "network", peerType, name, options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating WireGuard peer", err.Error())
		return
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

func (r *networkWireguardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state networkWireguardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	data, err := r.client.UCIGetAll(ctx, "network", name)
	if err != nil || len(data) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	r.optionsToModel(data, &state)
	state.ID = types.StringValue(fmt.Sprintf("network/%s", name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *networkWireguardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan networkWireguardModel
	var state networkWireguardModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	options := r.modelToOptions(ctx, plan)

	if err := r.client.UCITSet(ctx, "network", name, options); err != nil {
		resp.Diagnostics.AddError("Error updating WireGuard peer", err.Error())
		return
	}

	if err := r.client.UCICommit(ctx, "network"); err != nil {
		resp.Diagnostics.AddError("Error committing network config", err.Error())
		return
	}
	if err := r.client.UCIApply(ctx, false); err != nil {
		tflog.Warn(ctx, "Applying UCI changes failed", map[string]interface{}{"error": err.Error()})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *networkWireguardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state networkWireguardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	if err := r.client.UCIDelete(ctx, "network", name); err != nil {
		resp.Diagnostics.AddError("Error deleting WireGuard peer", err.Error())
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

func (r *networkWireguardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] != "network" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: network/<peer_name>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}

func (r *networkWireguardResource) modelToOptions(ctx context.Context, plan networkWireguardModel) map[string]interface{} {
	options := make(map[string]interface{})

	if !plan.Description.IsNull() {
		options["description"] = plan.Description.ValueString()
	}
	if !plan.PublicKey.IsNull() {
		options["public_key"] = plan.PublicKey.ValueString()
	}
	if !plan.EndpointHost.IsNull() {
		options["endpoint_host"] = plan.EndpointHost.ValueString()
	}
	if !plan.EndpointPort.IsNull() {
		options["endpoint_port"] = plan.EndpointPort.ValueInt64()
	}
	if !plan.PersistentKeepalive.IsNull() {
		options["persistent_keepalive"] = plan.PersistentKeepalive.ValueInt64()
	}
	if !plan.AllowedIPs.IsNull() {
		var ipsList []string
		diagnostics := plan.AllowedIPs.ElementsAs(ctx, &ipsList, false)
		if !diagnostics.HasError() {
			options["allowed_ips"] = strings.Join(ipsList, " ")
		}
	}

	return options
}

func (r *networkWireguardResource) optionsToModel(data map[string]interface{}, state *networkWireguardModel) {
	if v, ok := data["description"].(string); ok {
		state.Description = types.StringValue(v)
	}
	if v, ok := data["public_key"].(string); ok {
		state.PublicKey = types.StringValue(v)
	}
	if v, ok := data["endpoint_host"].(string); ok {
		state.EndpointHost = types.StringValue(v)
	}
	if v, ok := data["endpoint_port"]; ok {
		if f, ok := v.(float64); ok {
			state.EndpointPort = types.Int64Value(int64(f))
		}
	}
	if v, ok := data["persistent_keepalive"]; ok {
		if f, ok := v.(float64); ok {
			state.PersistentKeepalive = types.Int64Value(int64(f))
		}
	}
	if ips, ok := data["allowed_ips"]; ok {
		switch v := ips.(type) {
		case []interface{}:
			var ipsList []string
			for _, ip := range v {
				if s, ok := ip.(string); ok {
					ipsList = append(ipsList, s)
				}
			}
			state.AllowedIPs, _ = types.ListValueFrom(context.Background(), types.StringType, ipsList)
		case string:
			if v != "" {
				ipsList := strings.Split(v, " ")
				state.AllowedIPs, _ = types.ListValueFrom(context.Background(), types.StringType, ipsList)
			}
		}
	}
}

func extractInterfaceName(peerName string) string {
	parts := strings.SplitN(peerName, "_", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return peerName
}
