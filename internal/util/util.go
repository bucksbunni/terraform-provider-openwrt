package util

import "github.com/hashicorp/terraform-plugin-framework/types"

// helper to read string options
func StringOption(raw map[string]interface{}, key string) types.String {
	if v, ok := raw[key]; ok {
		if s, ok := v.(string); ok {
			return types.StringValue(s)
		}
	}
	return types.StringNull()
}

// helper: UCI "1"/"0" -> Bool
func BoolFromUCI(v interface{}) types.Bool {
	s, ok := v.(string)
	if !ok {
		return types.BoolNull()
	}
	switch s {
	case "1", "true", "yes", "on":
		return types.BoolValue(true)
	case "0", "false", "no", "off":
		return types.BoolValue(false)
	default:
		return types.BoolNull()
	}
}

// helper: Bool -> UCI "1"/"0"
func BoolToUCI(b types.Bool) (string, bool) {
	if b.IsNull() {
		return "", false
	}
	if b.ValueBool() {
		return "1", true
	}
	return "0", true
}
