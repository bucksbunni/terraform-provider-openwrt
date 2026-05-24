package converters

import "fmt"

// BoolToUCI converts a boolean to UCI string representation.
// UCI uses "1" for true and "0" for false.
func BoolToUCI(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// ParseUCIBool parses a UCI boolean value.
// Returns (value, isNull) where isNull is true if the value is nil or not a recognized boolean string.
// Recognized true values: "1", "true", "yes", "on"
// Recognized false values: "0", "false", "no", "off"
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

// ParseUCIInt parses a UCI integer value.
// Returns (value, isNull) where isNull is true if the value cannot be parsed.
// Supports float64, int, int64, and string types.
// String parsing supports negative numbers.
func ParseUCIInt(v interface{}) (int64, bool) {
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
		if len(n) > 0 && n[0] == '-' {
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

// StringPtr returns a pointer to the given string.
// Utility function for creating test data.
func StringPtr(s string) *string {
	return &s
}

// BoolPtr returns a pointer to the given bool.
// Utility function for creating test data.
func BoolPtr(b bool) *bool {
	return &b
}

// Int64Ptr returns a pointer to the given int64.
// Utility function for creating test data.
func Int64Ptr(i int64) *int64 {
	return &i
}

// MustParseUCIInt parses a UCI integer value, panicking on error.
// Use only in tests or when the value is guaranteed to be valid.
func MustParseUCIInt(v interface{}) int64 {
	val, null := ParseUCIInt(v)
	if null {
		panic(fmt.Sprintf("MustParseUCIInt received null value: %v", v))
	}
	return val
}
