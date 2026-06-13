package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewSysNetDevicesDataSource() datasource.DataSource {
	return &sysNetDevicesDataSource{}
}

type sysNetDevicesDataSource struct {
	client *JsonRpcClient
}

type sysNetDeviceModel struct {
	Name      types.String `tfsdk:"name"`
	MTU       types.Int64  `tfsdk:"mtu"`
	MACAddr   types.String `tfsdk:"macaddr"`
	IPAddr    types.String `tfsdk:"ipaddr"`
	IP6Addr   types.String `tfsdk:"ip6addr"`
	RXBytes   types.Int64  `tfsdk:"rx_bytes"`
	TXBytes   types.Int64  `tfsdk:"tx_bytes"`
	RXPackets types.Int64  `tfsdk:"rx_packets"`
	TXPackets types.Int64  `tfsdk:"tx_packets"`
	Flags     types.String `tfsdk:"flags"`
}

type sysNetDevicesModel struct {
	ID      types.String `tfsdk:"id"`
	Devices types.List   `tfsdk:"devices"`
}

func (d *sysNetDevicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sys_net_devices"
}

func (d *sysNetDevicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about network devices from OpenWrt.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:    true,
				Description: "Internal ID.",
			},
			"devices": dsschema.ListAttribute{
				Computed:    true,
				Description: "List of network devices.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":       types.StringType,
						"mtu":        types.Int64Type,
						"macaddr":    types.StringType,
						"ipaddr":     types.StringType,
						"ip6addr":    types.StringType,
						"rx_bytes":   types.Int64Type,
						"tx_bytes":   types.Int64Type,
						"rx_packets": types.Int64Type,
						"tx_packets": types.Int64Type,
						"flags":      types.StringType,
					},
				},
			},
		},
	}
}

func (d *sysNetDevicesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sysNetDevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider must be configured before using this data source.")
		return
	}

	// luci.sys.net.deviceinfo() was removed in modern LuCI. Gather per-device
	// details from sysfs and /proc/net/dev instead. Each output line is:
	//   name|mtu|macaddr|operstate|rx_bytes|rx_packets|tx_bytes|tx_packets
	const cmd = `for p in /sys/class/net/*; do n=${p##*/}; s=$(sed -n "s/^[ ]*$n://p" /proc/net/dev); set -- $s; echo "$n|$(cat $p/mtu 2>/dev/null)|$(cat $p/address 2>/dev/null)|$(cat $p/operstate 2>/dev/null)|$1|$2|$9|${10}"; done`

	out, err := d.client.SysExec(ctx, cmd)
	if err != nil {
		resp.Diagnostics.AddError("Error calling /rpc/sys", err.Error())
		return
	}

	lines := nonHeaderLines(out, 0)
	devices := make([]sysNetDeviceModel, 0, len(lines))
	for _, line := range lines {
		f := strings.Split(line, "|")
		if len(f) < 8 {
			continue
		}
		devices = append(devices, sysNetDeviceModel{
			Name:      types.StringValue(f[0]),
			MTU:       types.Int64Value(atoiSafe(f[1])),
			MACAddr:   types.StringValue(f[2]),
			Flags:     types.StringValue(f[3]),
			IPAddr:    types.StringValue(""),
			IP6Addr:   types.StringValue(""),
			RXBytes:   types.Int64Value(atoiSafe(f[4])),
			RXPackets: types.Int64Value(atoiSafe(f[5])),
			TXBytes:   types.Int64Value(atoiSafe(f[6])),
			TXPackets: types.Int64Value(atoiSafe(f[7])),
		})
	}

	devicesList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":       types.StringType,
			"mtu":        types.Int64Type,
			"macaddr":    types.StringType,
			"ipaddr":     types.StringType,
			"ip6addr":    types.StringType,
			"rx_bytes":   types.Int64Type,
			"tx_bytes":   types.Int64Type,
			"rx_packets": types.Int64Type,
			"tx_packets": types.Int64Type,
			"flags":      types.StringType,
		},
	}, devices)
	resp.Diagnostics.Append(diags...)

	state := sysNetDevicesModel{
		ID:      types.StringValue("net.deviceinfo"),
		Devices: devicesList,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
