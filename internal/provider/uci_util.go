package provider

// BoolToString converts a boolean to a UCI string representation.
// UCI uses "1" for true and "0" for false.
func BoolToString(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func ParseUCIBool(v interface{}) (bool, bool) {
	if v == nil {
		return false, true
	}
	s, ok := v.(string)
	if !ok {
		return false, true
	}
	switch s {
	case "1", "true", "yes", "on":
		return true, false
	case "0", "false", "no", "off":
		return false, false
	default:
		return false, true
	}
}

func StringFromUCI(raw map[string]interface{}, key string) (string, bool) {
	if raw == nil {
		return "", true
	}
	v, ok := raw[key]
	if !ok {
		return "", true
	}
	if v == nil {
		return "", true
	}
	s, ok := v.(string)
	if !ok {
		return "", true
	}
	return s, false
}

func Int64FromUCI(v interface{}) (int64, bool) {
	if v == nil {
		return 0, true
	}
	switch n := v.(type) {
	case float64:
		return int64(n), false
	case int:
		return int64(n), false
	case int64:
		return n, false
	case string:
		if n == "" {
			return 0, true
		}
		negative := false
		if n[0] == '-' {
			negative = true
			n = n[1:]
		}
		var i int64
		for _, c := range n {
			if c < '0' || c > '9' {
				return 0, true
			}
			i = i*10 + int64(c-'0')
		}
		if negative {
			i = -i
		}
		return i, false
	default:
		return 0, true
	}
}
