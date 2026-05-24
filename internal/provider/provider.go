package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// openwrtProvider implements the terraform-plugin-framework provider.Provider interface.
// It manages OpenWrt devices through the LuCI JSON-RPC API.
var _ provider.Provider = &openwrtProvider{}

// openwrtProvider is the main provider struct that holds configuration
// and provides access to the JSON-RPC client.
type openwrtProvider struct {
	version string
}

// NewProvider returns a function that creates a new OpenWrt provider instance.
// The version parameter is used for provider versioning information.
func NewProvider(version string) func() provider.Provider {
	return func() provider.Provider {
		return &openwrtProvider{
			version: version,
		}
	}
}

type providerModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func (p *openwrtProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "openwrt"
	resp.Version = p.version
}

func (p *openwrtProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required:    true,
				Description: "Base URL of the OpenWrt LuCI instance, e.g. http://192.168.1.1",
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "OpenWrt username (typically 'root').",
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "OpenWrt password.",
			},
			"insecure": schema.BoolAttribute{
				Optional:    true,
				Description: "Skip TLS certificate verification when using HTTPS.",
			},
		},
	}
}

func (p *openwrtProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get provider configuration
	host := config.Host.ValueString()
	user := config.Username.ValueString()
	pass := config.Password.ValueString()
	insecure := config.Insecure.ValueBool()

	// If variables are not available try to get them from the environment
	if host == "" {
		host = os.Getenv("OPENWRT_HOST")
	}
	if user == "" {
		user = os.Getenv("OPENWRT_USERNAME")
	}
	if pass == "" {
		pass = os.Getenv("OPENWRT_PASSWORD")
	}
	if os.Getenv("OPENWRT_INSECURE") != "" {
		insecure = os.Getenv("OPENWRT_INSECURE") == "true"
	}

	// Fail for missing credentials
	if host == "" || user == "" || pass == "" {
		resp.Diagnostics.AddError(
			"Missing configuration",
			"`host`, `username`, and `password` must be set.",
		)
		return
	}

	// Create a new client for the provider for operation
	client, err := NewJsonRpcClient(ctx, JsonRpcConfig{
		BaseURL:  host,
		Username: user,
		Password: pass,
		Insecure: insecure,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create OpenWrt JSON-RPC client",
			err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Configured OpenWrt provider", map[string]interface{}{
		"host": host,
	})

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *openwrtProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Generic
		NewUCISectionResource,
		NewFSFileResource,
		NewIPKGPackageResource,

		// Firewall
		NewFirewallRuleResource,
		NewFirewallZoneResource,
		NewFirewallForwardingResource,
		NewFirewallDefaultsResource,

		// Network
		NewNetworkInterfaceResource,
		NewNetworkDeviceResource,
		NewNetworkBridgeVlanResource,
		NewNetworkGlobalsResource,
		NewNetworkWireguardResource,

		// DHCP
		NewDHCPPoolResource,
		NewDHCPDNSMasqResource,
		NewDHCPOdhcpdResource,
		NewDHCPHostResource,

		// System
		NewSystemResource,
		NewSystemNTPResource,
		NewSystemLEDResource,

		// Dropbear (SSH)
		NewDropbearResource,

		// Wireless
		NewWirelessDeviceResource,
		NewWirelessInterfaceResource,

		// uHTTPd (Web Server)
		NewUHTTPdResource,
		NewUHTTPdCertResource,

		// RPC Daemon
		NewRPCDResource,
	}
}

func (p *openwrtProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Generic
		NewSysRPCDataSource,

		// System
		NewSysHostnameDataSource,
		NewSysUptimeDataSource,
		NewSysInitDataSource,

		// Network
		NewSysNetDevicesDataSource,
		NewSysNetRoutesDataSource,
		NewSysNetRoutes6DataSource,
		NewSysNetArpTableDataSource,
		NewSysNetConntrackDataSource,

		// Process
		NewSysProcessListDataSource,

		// Wireless
		NewSysWirelessDataSource,
	}
}

func NewStringValue(v string) types.String {
	return types.StringValue(v)
}

func NewBoolValue(v bool) types.Bool {
	return types.BoolValue(v)
}
