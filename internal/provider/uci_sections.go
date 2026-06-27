package provider

import (
	"context"
	"fmt"
	"strconv"
)

// This file centralizes how the provider addresses *anonymous* UCI sections.
//
// Background
//
// UCI sections come in two flavours. Named sections (for example `network.lan`)
// are addressed by their name. Anonymous sections (for example a `bridge-vlan`,
// a firewall `zone`, or a `dhcp` pool) have no user-facing name. The `uci` CLI
// renders them positionally as `network.@bridge-vlan[0]`, but that index is
// *not* stable: deleting an earlier section of the same type renumbers every
// section after it, so an index captured at create time can silently point at
// the wrong section later. What libuci does guarantee is a stable internal
// identifier - the `.name` key, e.g. `cfg0abc12` - that lives for as long as
// the section does. That identifier, not the positional index, is the correct
// handle for create/read/update/delete over the LuCI JSON-RPC API.
//
// Strategy
//
// Each resource backed by an anonymous section persists that `.name` in a
// Computed `section` attribute at Create time (taken from UCIAdd) and
// addresses the section directly afterwards. UCIResolveSection encapsulates the
// lookup: it trusts the persisted identifier when present and falls back to
// matching on stable option values (such as a name, or a device+vlan pair) only
// when the identifier is missing or stale - for example right after a
// `terraform import`, or if the section was recreated out of band. Treating a
// missing section as "already deleted" lives here too, so destroy is idempotent
// across every anonymous-section resource instead of being re-implemented (and
// subtly diverging) in each one.

// uciOptionString normalizes a UCI option value returned by the LuCI JSON-RPC
// into the string form used for comparisons. UCI is itself untyped - every
// value is stored on disk as a string - but the JSON-RPC layer may surface a
// value as a JSON number or boolean depending on the call, so each of those is
// coerced back to its canonical string here. A nil value (absent option) maps
// to the empty string.
func uciOptionString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "1"
		}
		return "0"
	case float64:
		// JSON numbers decode as float64. Render whole numbers without a
		// trailing ".0" so a vlan id like 10 compares equal to the string "10".
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", t)
	}
}

// uciSectionMatches reports whether a section table - as returned by UCIForeach
// or UCIGetAll - has every option named in match set to the corresponding
// value. Comparison is done on the string form of each value via
// uciOptionString, so callers pass numeric keys (such as a vlan id) as their
// decimal string. An empty match map matches every section.
func uciSectionMatches(section map[string]interface{}, match map[string]string) bool {
	for option, want := range match {
		if uciOptionString(section[option]) != want {
			return false
		}
	}
	return true
}

// UCIResolveSection locates a single anonymous UCI section and returns its
// option table together with its stable internal identifier (the `.name` key,
// e.g. `cfg0abc12`).
//
// It resolves the identifier in two steps:
//
//  1. If knownName is non-empty - the identifier previously persisted in
//     Terraform state by Create or Read - the section is fetched directly with
//     a single uci.get_all call. This is the steady-state path: it avoids
//     scanning every section of the type, and the match map is still checked so
//     a stale identifier is rejected rather than trusted blindly.
//  2. Otherwise, or if the direct lookup did not yield a matching section, every
//     section of typ is scanned and the first one whose options satisfy match is
//     returned. This covers a freshly imported resource (no identifier in state
//     yet) and the rare case of an out-of-band recreation under a new identifier.
//
// found is false when the section cannot be located at all. Callers treat that
// as "already deleted" - removing the resource from state on Read/Delete - which
// is what keeps destroy idempotent even when a parent resource removed the
// section first.
func (c *JsonRpcClient) UCIResolveSection(ctx context.Context, config, typ, knownName string, match map[string]string) (map[string]interface{}, string, bool, error) {
	if knownName != "" {
		data, err := c.UCIGetAll(ctx, config, knownName)
		if err == nil && len(data) > 0 && uciSectionMatches(data, match) {
			return data, knownName, true, nil
		}
	}

	sections, err := c.UCIForeach(ctx, config, typ)
	if err != nil {
		return nil, "", false, err
	}
	for _, s := range sections {
		if uciSectionMatches(s, match) {
			name, _ := s[".name"].(string)
			return s, name, true, nil
		}
	}
	return nil, "", false, nil
}
