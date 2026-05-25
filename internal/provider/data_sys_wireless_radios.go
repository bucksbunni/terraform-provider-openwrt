package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewSysWirelessRadiosDataSource() datasource.DataSource {
	return &sysWirelessRadiosDataSource{}
}

type sysWirelessRadiosDataSource struct {
	client *JsonRpcClient
}

type sysWirelessRadioModel struct {
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Path     types.String `tfsdk:"path"`
	Band     types.String `tfsdk:"band"`
	Channel  types.Int64  `tfsdk:"channel"`
	HTMode   types.String `tfsdk:"htmode"`
	HWMode   types.String `tfsdk:"hwmode"`
	Country  types.String `tfsdk:"country"`
	TxPower  types.Int64  `tfsdk:"txpower"`
	Disabled types.Bool   `tfsdk:"disabled"`
}

type sysWirelessRadiosModel struct {
	ID     types.String `tfsdk:"id"`
	Radios types.List   `tfsdk:"radios"`
}

func (d *sysWirelessRadiosDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sys_wireless_radios"
}

func (d *sysWirelessRadiosDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Discovers available wireless radio devices from OpenWrt.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:    true,
				Description: "Internal ID.",
			},
			"radios": dsschema.ListAttribute{
				Computed:    true,
				Description: "List of wireless radio devices.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":      types.StringType,
						"type":      types.StringType,
						"path":      types.StringType,
						"band":      types.StringType,
						"channel":   types.Int64Type,
						"htmode":    types.StringType,
						"hwmode":    types.StringType,
						"country":   types.StringType,
						"txpower":   types.Int64Type,
						"disabled":  types.BoolType,
					},
				},
			},
		},
	}
}

func (d *sysWirelessRadiosDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sysWirelessRadiosDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider must be configured before using this data source.")
		return
	}

	radios, err := d.client.UCIForeach(ctx, "wireless", "wifi-device")
	if err != nil {
		resp.Diagnostics.AddError("Error reading wireless radios", err.Error())
		return
	}

	radioList := make([]sysWirelessRadioModel, 0, len(radios))
	for _, radio := range radios {
		r := sysWirelessRadioModel{}

		if v, ok := radio[".name"].(string); ok {
			r.Name = types.StringValue(v)
		}
		if v, ok := radio["type"].(string); ok {
			r.Type = types.StringValue(v)
		}
		if v, ok := radio["path"].(string); ok {
			r.Path = types.StringValue(v)
		}
		if v, ok := radio["band"].(string); ok {
			r.Band = types.StringValue(v)
		}
		if v, ok := radio["channel"]; ok {
			if f, ok := v.(float64); ok {
				r.Channel = types.Int64Value(int64(f))
			}
		}
		if v, ok := radio["htmode"].(string); ok {
			r.HTMode = types.StringValue(v)
		}
		if v, ok := radio["hwmode"].(string); ok {
			r.HWMode = types.StringValue(v)
		}
		if v, ok := radio["country"].(string); ok {
			r.Country = types.StringValue(v)
		}
		if v, ok := radio["txpower"]; ok {
			if f, ok := v.(float64); ok {
				r.TxPower = types.Int64Value(int64(f))
			}
		}
		if v, ok := radio["disabled"].(string); ok {
			r.Disabled = types.BoolValue(v == "1" || v == "true")
		}

		radioList = append(radioList, r)
	}

	radiosList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":     types.StringType,
			"type":     types.StringType,
			"path":     types.StringType,
			"band":     types.StringType,
			"channel":  types.Int64Type,
			"htmode":   types.StringType,
			"hwmode":   types.StringType,
			"country":  types.StringType,
			"txpower":  types.Int64Type,
			"disabled": types.BoolType,
		},
	}, radioList)
	resp.Diagnostics.Append(diags...)

	state := sysWirelessRadiosModel{
		ID:     types.StringValue("wireless_radios"),
		Radios: radiosList,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}