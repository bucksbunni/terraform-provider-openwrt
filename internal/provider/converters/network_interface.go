package converters

// NetworkInterfaceOptions represents network interface configuration options.
// All fields are pointers to allow distinguishing between unset and empty values.
type NetworkInterfaceOptions struct {
	Proto       *string
	Device      *string
	IPAddr      *string
	Netmask     *string
	Gateway     *string
	DNS         *string
	Metric      *int64
	Delegate    *bool
	IP6Addr     *string
	IP6Prefix   *string
	IP6Assign   *string
	IP6Gateway  *string
	Auto        *string
	IfType      *string
	BridgeEmpty *bool
}

// NetworkInterfaceToOptions converts NetworkInterfaceOptions to a UCI options map.
// Pointer fields that are nil are omitted from the result.
func NetworkInterfaceToOptions(m NetworkInterfaceOptions) map[string]interface{} {
	options := make(map[string]interface{})

	if m.Proto != nil {
		options["proto"] = *m.Proto
	}
	if m.Device != nil {
		options["device"] = *m.Device
	}
	if m.IPAddr != nil {
		options["ipaddr"] = *m.IPAddr
	}
	if m.Netmask != nil {
		options["netmask"] = *m.Netmask
	}
	if m.Gateway != nil {
		options["gateway"] = *m.Gateway
	}
	if m.DNS != nil {
		options["dns"] = *m.DNS
	}
	if m.Metric != nil {
		options["metric"] = *m.Metric
	}
	if m.Delegate != nil {
		options["delegate"] = BoolToUCI(*m.Delegate)
	}
	if m.IP6Addr != nil {
		options["ip6addr"] = *m.IP6Addr
	}
	if m.IP6Prefix != nil {
		options["ip6prefix"] = *m.IP6Prefix
	}
	if m.IP6Assign != nil {
		options["ip6assign"] = *m.IP6Assign
	}
	if m.IP6Gateway != nil {
		options["ip6gateway"] = *m.IP6Gateway
	}
	if m.Auto != nil {
		options["auto"] = *m.Auto
	}
	if m.IfType != nil {
		options["type"] = *m.IfType
	}
	if m.BridgeEmpty != nil {
		options["bridge_empty"] = BoolToUCI(*m.BridgeEmpty)
	}

	return options
}

// OptionsToNetworkInterface converts a UCI options map to NetworkInterfaceOptions.
func OptionsToNetworkInterface(data map[string]interface{}) NetworkInterfaceOptions {
	m := NetworkInterfaceOptions{}

	if v, ok := data["proto"].(string); ok {
		m.Proto = &v
	}
	if v, ok := data["device"].(string); ok {
		m.Device = &v
	}
	if v, ok := data["ipaddr"].(string); ok {
		m.IPAddr = &v
	}
	if v, ok := data["netmask"].(string); ok {
		m.Netmask = &v
	}
	if v, ok := data["gateway"].(string); ok {
		m.Gateway = &v
	}
	if v, ok := data["dns"].(string); ok {
		m.DNS = &v
	}
	if v, ok := data["metric"]; ok {
		if val, null := ParseUCIInt(v); !null {
			m.Metric = &val
		}
	}
	if v, ok := data["delegate"].(string); ok {
		if val, null := ParseUCIBool(v); !null {
			m.Delegate = &val
		}
	}
	if v, ok := data["ip6addr"].(string); ok {
		m.IP6Addr = &v
	}
	if v, ok := data["ip6prefix"].(string); ok {
		m.IP6Prefix = &v
	}
	if v, ok := data["ip6assign"].(string); ok {
		m.IP6Assign = &v
	}
	if v, ok := data["ip6gateway"].(string); ok {
		m.IP6Gateway = &v
	}
	if v, ok := data["auto"].(string); ok {
		m.Auto = &v
	}
	if v, ok := data["type"].(string); ok {
		m.IfType = &v
	}
	if v, ok := data["bridge_empty"].(string); ok {
		if val, null := ParseUCIBool(v); !null {
			m.BridgeEmpty = &val
		}
	}

	return m
}
