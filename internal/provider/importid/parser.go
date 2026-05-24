package importid

import "fmt"

// ParseNetworkInterfaceID parses a network interface import ID and returns the interface name.
// The ID format is "network/<interface_name>".
func ParseNetworkInterfaceID(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("network interface ID cannot be empty")
	}
	config, name, err := ParseUCISectionID(id)
	if err != nil {
		return "", err
	}
	if config != "network" {
		return "", fmt.Errorf("expected config 'network', got %q", config)
	}
	return name, nil
}

func FormatNetworkInterfaceID(name string) string {
	return FormatUCISectionID("network", name)
}

func ParseDHCPPoolID(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("DHCP pool ID cannot be empty")
	}
	config, name, err := ParseUCISectionID(id)
	if err != nil {
		return "", err
	}
	if config != "dhcp" {
		return "", fmt.Errorf("expected config 'dhcp', got %q", config)
	}
	return name, nil
}

func FormatDHCPPoolID(name string) string {
	return FormatUCISectionID("dhcp", name)
}

func ParseDHCPHostID(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("DHCP host ID cannot be empty")
	}
	config, name, err := ParseUCISectionID(id)
	if err != nil {
		return "", err
	}
	if config != "dhcp" {
		return "", fmt.Errorf("expected config 'dhcp', got %q", config)
	}
	return name, nil
}

func FormatDHCPHostID(name string) string {
	return FormatUCISectionID("dhcp", name)
}

func ParseFirewallZoneID(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("firewall zone ID cannot be empty")
	}
	config, name, err := ParseUCISectionID(id)
	if err != nil {
		return "", err
	}
	if config != "firewall" {
		return "", fmt.Errorf("expected config 'firewall', got %q", config)
	}
	return name, nil
}

func FormatFirewallZoneID(name string) string {
	return FormatUCISectionID("firewall", name)
}

func ParseFirewallRuleID(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("firewall rule ID cannot be empty")
	}
	config, name, err := ParseUCISectionID(id)
	if err != nil {
		return "", err
	}
	if config != "firewall" {
		return "", fmt.Errorf("expected config 'firewall', got %q", config)
	}
	return name, nil
}

func FormatFirewallRuleID(name string) string {
	return FormatUCISectionID("firewall", name)
}

func ParseFirewallForwardingID(id string) (src, dest string, err error) {
	if id == "" {
		return "", "", fmt.Errorf("firewall forwarding ID cannot be empty")
	}
	config, name, err := ParseUCISectionID(id)
	if err != nil {
		return "", "", err
	}
	if config != "firewall" {
		return "", "", fmt.Errorf("expected config 'firewall', got %q", config)
	}
	// Forwarding IDs are typically in format "src_dest"
	for i := 0; i < len(name); i++ {
		if name[i] == '_' && i > 0 {
			return name[:i], name[i+1:], nil
		}
	}
	return "", "", fmt.Errorf("invalid firewall forwarding ID format: %q", id)
}

func FormatFirewallForwardingID(src, dest string) string {
	return FormatUCISectionID("firewall", src+"_"+dest)
}

func ParseWirelessInterfaceID(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("wireless interface ID cannot be empty")
	}
	config, name, err := ParseUCISectionID(id)
	if err != nil {
		return "", err
	}
	if config != "wireless" {
		return "", fmt.Errorf("expected config 'wirewall', got %q", config)
	}
	return name, nil
}

func FormatWirelessInterfaceID(name string) string {
	return FormatUCISectionID("wireless", name)
}

func ParseDropbearID(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("dropbear ID cannot be empty")
	}
	config, name, err := ParseUCISectionID(id)
	if err != nil {
		return "", err
	}
	if config != "dropbear" {
		return "", fmt.Errorf("expected config 'dropbear', got %q", config)
	}
	return name, nil
}

func FormatDropbearID(name string) string {
	return FormatUCISectionID("dropbear", name)
}

func ParseSystemLEDID(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("system LED ID cannot be empty")
	}
	config, name, err := ParseUCISectionID(id)
	if err != nil {
		return "", err
	}
	if config != "system" {
		return "", fmt.Errorf("expected config 'system', got %q", config)
	}
	return name, nil
}

func FormatSystemLEDID(name string) string {
	return FormatUCISectionID("system", name)
}

func ParseUCISectionID(id string) (config, name string, err error) {
	if id == "" {
		return "", "", fmt.Errorf("UCI section ID cannot be empty")
	}
	slashCount := 0
	lastSlash := -1
	for i := 0; i < len(id); i++ {
		if id[i] == '/' {
			slashCount++
			lastSlash = i
			if i == 0 || i == len(id)-1 {
				return "", "", fmt.Errorf("invalid UCI section ID format: %q", id)
			}
		}
	}
	if slashCount != 1 {
		return "", "", fmt.Errorf("invalid UCI section ID format: %q (expected exactly one '/')", id)
	}
	config = id[:lastSlash]
	name = id[lastSlash+1:]
	return config, name, nil
}

func FormatUCISectionID(config, name string) string {
	return config + "/" + name
}
