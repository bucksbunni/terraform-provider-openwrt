package provider

import (
	"encoding/hex"
	"net"
	"strconv"
	"strings"
)

// Helpers for parsing kernel /proc tables. Modern OpenWrt (LuCI 24.10+) dropped
// the luci.sys.net.{routes,routes6,arptable,deviceinfo} RPC methods, so the
// corresponding data sources read /proc directly via the sys.exec RPC.

// hexLEToIPv4 converts a little-endian hex word (as found in /proc/net/route,
// e.g. "0102300A") to a dotted-quad string ("10.48.2.1"). It returns "" on
// malformed input.
func hexLEToIPv4(s string) string {
	v, err := strconv.ParseUint(strings.TrimSpace(s), 16, 32)
	if err != nil {
		return ""
	}
	return strconv.FormatUint(v&0xff, 10) + "." +
		strconv.FormatUint((v>>8)&0xff, 10) + "." +
		strconv.FormatUint((v>>16)&0xff, 10) + "." +
		strconv.FormatUint((v>>24)&0xff, 10)
}

// hexToIPv6 converts a 32-character hex string (as found in /proc/net/ipv6_route)
// to its canonical IPv6 text form. It returns "" on malformed input.
func hexToIPv6(s string) string {
	b, err := hex.DecodeString(strings.TrimSpace(s))
	if err != nil || len(b) != net.IPv6len {
		return ""
	}
	return net.IP(b).String()
}

// hexToInt parses a hexadecimal integer, returning 0 on malformed input.
func hexToInt(s string) int64 {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 16, 64)
	if err != nil {
		return 0
	}
	return v
}

// atoiSafe parses a decimal integer, returning 0 on malformed input.
func atoiSafe(s string) int64 {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0
	}
	return v
}

// toInt64 extracts an int64 from a JSON-decoded value that may be a number or a
// numeric string (LuCI's sys RPCs are inconsistent about which they use).
func toInt64(v interface{}) (int64, bool) {
	switch t := v.(type) {
	case float64:
		return int64(t), true
	case string:
		if n, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64); err == nil {
			return n, true
		}
	}
	return 0, false
}

// nonHeaderLines splits exec output into trimmed, non-empty lines, dropping the
// first headerLines lines.
func nonHeaderLines(out string, headerLines int) []string {
	var lines []string
	for i, l := range strings.Split(out, "\n") {
		if i < headerLines {
			continue
		}
		if strings.TrimSpace(l) == "" {
			continue
		}
		lines = append(lines, l)
	}
	return lines
}
