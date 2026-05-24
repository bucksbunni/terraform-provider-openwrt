package converters

// SystemLEDOptions represents a system LED configuration for OpenWrt's UCI system config.
// These options are used when creating or updating LED configurations via the UCI RPC API.
type SystemLEDOptions struct {
	// Name is the LED name displayed in LuCI.
	Name *string
	// SysFS specifies the LED device path (e.g., "apu:green:3", "led0").
	SysFS *string
	// Trigger specifies the trigger type: "none", "netdev", "timer", "heartbeat", "gpio", "usbdev".
	Trigger *string
	// Mode specifies trigger-specific modes (e.g., "link tx rx" for netdev trigger).
	Mode *string
	// Dev specifies the network device for netdev trigger.
	Dev *string
	// Default controls whether the LED is enabled by default.
	Default *bool
}

// SystemLEDToOptions converts a SystemLEDOptions struct to UCI options map.
// Nil pointer fields are omitted from the result.
func SystemLEDToOptions(m SystemLEDOptions) map[string]interface{} {
	options := make(map[string]interface{})

	if m.Name != nil {
		options["name"] = *m.Name
	}
	if m.SysFS != nil {
		options["sysfs"] = *m.SysFS
	}
	if m.Trigger != nil {
		options["trigger"] = *m.Trigger
	}
	if m.Mode != nil {
		options["mode"] = *m.Mode
	}
	if m.Dev != nil {
		options["dev"] = *m.Dev
	}
	if m.Default != nil && *m.Default {
		options["default"] = "1"
	}

	return options
}

// OptionsToSystemLED converts UCI data map to SystemLEDOptions struct.
// Fields that are missing or invalid in the UCI data are left as nil pointers.
func OptionsToSystemLED(data map[string]interface{}) SystemLEDOptions {
	m := SystemLEDOptions{}

	if v, ok := data["name"].(string); ok {
		m.Name = &v
	}
	if v, ok := data["sysfs"].(string); ok {
		m.SysFS = &v
	}
	if v, ok := data["trigger"].(string); ok {
		m.Trigger = &v
	}
	if v, ok := data["mode"].(string); ok {
		m.Mode = &v
	}
	if v, ok := data["dev"].(string); ok {
		m.Dev = &v
	}
	if v, ok := data["default"].(string); ok {
		if val, null := ParseUCIBool(v); !null {
			m.Default = &val
		}
	}

	return m
}
