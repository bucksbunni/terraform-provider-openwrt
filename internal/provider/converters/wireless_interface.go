package converters

// WirelessInterfaceOptions represents a wireless interface configuration for OpenWrt's UCI wireless config.
// These options are used when creating or updating wireless interfaces via the UCI RPC API.
type WirelessInterfaceOptions struct {
	// Name is the wireless interface name (e.g., "wifinet0", "wifinet1").
	Name *string
	// Device specifies the radio device (e.g., "radio0").
	Device *string
	// Mode sets the operation mode: "ap", "sta", "adhoc", "monitor", "wds".
	Mode *string
	// SSID is the wireless network name.
	SSID *string
	// Encryption specifies the encryption mode: "psk", "psk2", "psk2+ccmp", "sae", "wpa3", etc.
	Encryption *string
	// Key is the encryption key/passphrase.
	Key *string
	// Network specifies the associated network interface.
	Network *string
	// Disabled controls whether the interface is disabled.
	Disabled *bool
	// Hidden controls whether the SSID is hidden.
	Hidden *bool
	// MACFilter sets MAC filtering mode: "disable", "allow", "deny".
	MACFilter *string
	// MACList specifies MAC addresses for filtering (space-separated).
	MACList *string
	// Isolate controls client isolation.
	Isolate *bool
}

// WirelessInterfaceToOptions converts a WirelessInterfaceOptions struct to UCI options map.
// Nil pointer fields are omitted from the result.
// Note: Disabled=true and Hidden=true are represented by "1", Disabled=false and Hidden=false are omitted.
func WirelessInterfaceToOptions(m WirelessInterfaceOptions) map[string]interface{} {
	options := make(map[string]interface{})

	if m.Name != nil {
		options["name"] = *m.Name
	}
	if m.Device != nil {
		options["device"] = *m.Device
	}
	if m.Mode != nil {
		options["mode"] = *m.Mode
	}
	if m.SSID != nil {
		options["ssid"] = *m.SSID
	}
	if m.Encryption != nil {
		options["encryption"] = *m.Encryption
	}
	if m.Key != nil {
		options["key"] = *m.Key
	}
	if m.Network != nil {
		options["network"] = *m.Network
	}
	if m.Disabled != nil && *m.Disabled {
		options["disabled"] = "1"
	}
	if m.Hidden != nil && *m.Hidden {
		options["hidden"] = "1"
	}
	if m.MACFilter != nil {
		options["macfilter"] = *m.MACFilter
	}
	if m.MACList != nil {
		options["maclist"] = *m.MACList
	}
	if m.Isolate != nil && *m.Isolate {
		options["isolate"] = "1"
	}

	return options
}

// OptionsToWirelessInterface converts UCI data map to WirelessInterfaceOptions struct.
// Fields that are missing or invalid in the UCI data are left as nil pointers.
func OptionsToWirelessInterface(data map[string]interface{}) WirelessInterfaceOptions {
	m := WirelessInterfaceOptions{}

	if v, ok := data["name"].(string); ok {
		m.Name = &v
	}
	if v, ok := data["device"].(string); ok {
		m.Device = &v
	}
	if v, ok := data["mode"].(string); ok {
		m.Mode = &v
	}
	if v, ok := data["ssid"].(string); ok {
		m.SSID = &v
	}
	if v, ok := data["encryption"].(string); ok {
		m.Encryption = &v
	}
	if v, ok := data["key"].(string); ok {
		m.Key = &v
	}
	if v, ok := data["network"].(string); ok {
		m.Network = &v
	}
	if v, ok := data["disabled"].(string); ok {
		if val, null := ParseUCIBool(v); !null {
			m.Disabled = &val
		}
	}
	if v, ok := data["hidden"].(string); ok {
		if val, null := ParseUCIBool(v); !null {
			m.Hidden = &val
		}
	}
	if v, ok := data["macfilter"].(string); ok {
		m.MACFilter = &v
	}
	if v, ok := data["maclist"].(string); ok {
		m.MACList = &v
	}
	if v, ok := data["isolate"].(string); ok {
		if val, null := ParseUCIBool(v); !null {
			m.Isolate = &val
		}
	}

	return m
}
