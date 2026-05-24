package provider

import "testing"

func TestBoolToString(t *testing.T) {
	tests := []struct {
		name string
		b    bool
		want string
	}{
		{"true", true, "1"},
		{"false", false, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BoolToString(tt.b); got != tt.want {
				t.Errorf("BoolToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseUCIBool(t *testing.T) {
	tests := []struct {
		name     string
		v        interface{}
		wantVal  bool
		wantNull bool
	}{
		{"nil", nil, false, true},
		{"string 1", "1", true, false},
		{"string true", "true", true, false},
		{"string yes", "yes", true, false},
		{"string on", "on", true, false},
		{"string 0", "0", false, false},
		{"string false", "false", false, false},
		{"string no", "no", false, false},
		{"string off", "off", false, false},
		{"string invalid", "invalid", false, true},
		{"string empty", "", false, true},
		{"int 1", 1, false, true},
		{"bool true", true, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotNull := ParseUCIBool(tt.v)
			if gotVal != tt.wantVal {
				t.Errorf("ParseUCIBool() val = %v, want %v", gotVal, tt.wantVal)
			}
			if gotNull != tt.wantNull {
				t.Errorf("ParseUCIBool() null = %v, want %v", gotNull, tt.wantNull)
			}
		})
	}
}

func TestStringFromUCI(t *testing.T) {
	tests := []struct {
		name     string
		raw      map[string]interface{}
		key      string
		wantVal  string
		wantNull bool
	}{
		{"existing key", map[string]interface{}{"key": "value"}, "key", "value", false},
		{"missing key", map[string]interface{}{}, "key", "", true},
		{"nil map", nil, "key", "", true},
		{"nil value", map[string]interface{}{"key": nil}, "key", "", true},
		{"non-string value", map[string]interface{}{"key": 123}, "key", "", true},
		{"empty string", map[string]interface{}{"key": ""}, "key", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotNull := StringFromUCI(tt.raw, tt.key)
			if gotVal != tt.wantVal {
				t.Errorf("StringFromUCI() val = %v, want %v", gotVal, tt.wantVal)
			}
			if gotNull != tt.wantNull {
				t.Errorf("StringFromUCI() null = %v, want %v", gotNull, tt.wantNull)
			}
		})
	}
}

func TestInt64FromUCI(t *testing.T) {
	tests := []struct {
		name     string
		v        interface{}
		wantVal  int64
		wantNull bool
	}{
		{"nil", nil, 0, true},
		{"float64 42", float64(42), 42, false},
		{"int 42", 42, 42, false},
		{"int64 42", int64(42), 42, false},
		{"string 42", "42", 42, false},
		{"string 0", "0", 0, false},
		{"string negative", "-5", -5, false},
		{"string empty", "", 0, true},
		{"string invalid", "abc", 0, true},
		{"string with spaces", " 42", 0, true},
		{"bool true", true, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotNull := Int64FromUCI(tt.v)
			if gotVal != tt.wantVal {
				t.Errorf("Int64FromUCI() val = %v, want %v", gotVal, tt.wantVal)
			}
			if gotNull != tt.wantNull {
				t.Errorf("Int64FromUCI() null = %v, want %v", gotNull, tt.wantNull)
			}
		})
	}
}
