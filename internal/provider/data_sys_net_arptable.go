package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewSysNetArpTableDataSource() datasource.DataSource {
	return &sysNetArpTableDataSource{}
}

type sysNetArpTableDataSource struct {
	client *JsonRpcClient
}

type sysNetArpEntryModel struct {
	IPAddress types.String `tfsdk:"ip_address"`
	HWAddress types.String `tfsdk:"hw_address"`
	HWType    types.String `tfsdk:"hw_type"`
	Flags     types.String `tfsdk:"flags"`
	Mask      types.String `tfsdk:"mask"`
	Device    types.String `tfsdk:"device"`
}

type sysNetArpTableModel struct {
	ID      types.String `tfsdk:"id"`
	Entries types.List   `tfsdk:"entries"`
}

func (d *sysNetArpTableDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sys_net_arptable"
}

func (d *sysNetArpTableDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves the ARP table from OpenWrt.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:    true,
				Description: "Internal ID.",
			},
			"entries": dsschema.ListAttribute{
				Computed:    true,
				Description: "List of ARP entries.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"ip_address": types.StringType,
						"hw_address": types.StringType,
						"hw_type":    types.StringType,
						"flags":      types.StringType,
						"mask":       types.StringType,
						"device":     types.StringType,
					},
				},
			},
		},
	}
}

func (d *sysNetArpTableDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*JsonRpcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			"Expected *JsonRpcClient",
		)
		return
	}
	d.client = client
}

func (d *sysNetArpTableDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider must be configured before using this data source.")
		return
	}

	// luci.sys.net.arptable() was removed in modern LuCI, so read the kernel
	// ARP cache directly. /proc/net/arp columns:
	//   IP address   HW type   Flags   HW address   Mask   Device
	out, err := d.client.SysExec(ctx, "cat /proc/net/arp")
	if err != nil {
		resp.Diagnostics.AddError("Error calling /rpc/sys", err.Error())
		return
	}

	lines := nonHeaderLines(out, 1)
	entries := make([]sysNetArpEntryModel, 0, len(lines))
	for _, line := range lines {
		f := strings.Fields(line)
		if len(f) < 6 {
			continue
		}
		entries = append(entries, sysNetArpEntryModel{
			IPAddress: types.StringValue(f[0]),
			HWType:    types.StringValue(f[1]),
			Flags:     types.StringValue(f[2]),
			HWAddress: types.StringValue(f[3]),
			Mask:      types.StringValue(f[4]),
			Device:    types.StringValue(f[5]),
		})
	}

	entriesList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"ip_address": types.StringType,
			"hw_address": types.StringType,
			"hw_type":    types.StringType,
			"flags":      types.StringType,
			"mask":       types.StringType,
			"device":     types.StringType,
		},
	}, entries)
	resp.Diagnostics.Append(diags...)

	state := sysNetArpTableModel{
		ID:      types.StringValue("net.arptable"),
		Entries: entriesList,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
