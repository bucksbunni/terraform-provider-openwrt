package provider

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &fsFileResource{}
var _ resource.ResourceWithImportState = &fsFileResource{}

func NewFSFileResource() resource.Resource {
	return &fsFileResource{}
}

type fsFileResource struct {
	client *JsonRpcClient
}

type fsFileModel struct {
	ID      types.String `tfsdk:"id"`
	Path    types.String `tfsdk:"path"`
	Content types.String `tfsdk:"content"`
	// You can add mode, owner, group as needed and use fs.chmod/fs.chown.
}

func (r *fsFileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fs_file"
}

func (r *fsFileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages a file on the OpenWrt filesystem via LuCI /rpc/fs.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:    true,
				Description: "Internal ID, equal to path.",
			},
			"path": rschema.StringAttribute{
				Required:    true,
				Description: "Absolute path to the file on the router.",
			},
			"content": rschema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "File content as UTF-8 string.",
			},
		},
	}
}

func (r *fsFileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*JsonRpcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *JsonRpcClient, got %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *fsFileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan fsFileModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := plan.Path.ValueString()
	content := plan.Content.ValueString()

	b64 := base64.StdEncoding.EncodeToString([]byte(content))
	if err := r.client.FSWriteFile(ctx, path, b64); err != nil {
		resp.Diagnostics.AddError("Error writing file", err.Error())
		return
	}

	plan.ID = types.StringValue(path)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *fsFileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state fsFileModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := state.Path.ValueString()

	// If file is gone, drop from state.
	_, err := r.client.FSStat(ctx, path)
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	b64, err := r.client.FSReadFile(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError("Error reading file", err.Error())
		return
	}
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		resp.Diagnostics.AddError("Error decoding file content", err.Error())
		return
	}

	state.Content = types.StringValue(string(raw))
	state.ID = types.StringValue(path)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *fsFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan fsFileModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := plan.Path.ValueString()
	content := plan.Content.ValueString()

	b64 := base64.StdEncoding.EncodeToString([]byte(content))
	if err := r.client.FSWriteFile(ctx, path, b64); err != nil {
		resp.Diagnostics.AddError("Error writing file", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *fsFileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state fsFileModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := state.Path.ValueString()

	if err := r.client.FSUnlink(ctx, path); err != nil {
		// If it already doesn't exist, ignore.
	}

	resp.State.RemoveResource(ctx)
}

func (r *fsFileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// ID is full path
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("path"), req.ID)...)
}
