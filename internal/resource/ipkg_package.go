package resource

import (
	"context"
	"fmt"

	"github.com/bucksbunni/terraform-provider-openwrt/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ipkgPackageResource{}
var _ resource.ResourceWithImportState = &ipkgPackageResource{}

func NewIPKGPackageResource() resource.Resource {
	return &ipkgPackageResource{}
}

type ipkgPackageResource struct {
	client *client.JsonRpcClient
}

type ipkgPackageModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (r *ipkgPackageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipkg_package"
}

func (r *ipkgPackageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenWrt package via LuCI /rpc/ipkg.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:    true,
				Description: "Internal ID, equal to package name.",
			},
			"name": rschema.StringAttribute{
				Required:    true,
				Description: "Package name as known to opkg, e.g. 'luci-mod-rpc'.",
			},
		},
	}
}

func (r *ipkgPackageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.JsonRpcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.JsonRpcClient, got %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *ipkgPackageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ipkgPackageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()

	if err := r.client.IPKGInstall(ctx, name); err != nil {
		resp.Diagnostics.AddError("Error installing package", err.Error())
		return
	}

	plan.ID = types.StringValue(name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ipkgPackageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ipkgPackageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	ok, err := r.client.IPKGInstalled(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error checking package installation", err.Error())
		return
	}

	if !ok {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ipkgPackageResource) Update(ctx context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Package name is immutable; nothing to do.
	resp.Diagnostics.AddWarning("No-op update", "Package name cannot be changed; recreate the resource instead.")
}

func (r *ipkgPackageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ipkgPackageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	if err := r.client.IPKGRemove(ctx, name); err != nil {
		resp.Diagnostics.AddError("Error removing package", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *ipkgPackageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// ID and name are identical
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}
