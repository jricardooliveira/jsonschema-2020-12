package formats

import (
	"strings"
	"testing"
)

// RFC 1123 2.1.
func TestHostname(t *testing.T) {
	table(t, "hostname", map[string]bool{
		"example.com": true, "a": true, "a-b.c1": true, "1.2": true,
		strings.Repeat("a", 63) + ".com": true,
		"example.com.":                   false, "": false, "-a.com": false, "a-.com": false,
		"a..b": false, ".a": false, "a_b.com": false, "a b": false,
		strings.Repeat("a", 64) + ".com": false,
		strings.Repeat("a.", 127):        false,
	})
	// Total length cap alone: labels are within 63, no trailing dot.
	l := strings.Repeat("a", 63) + "."
	pre := l + l + l // 192 bytes
	table(t, "hostname", map[string]bool{
		pre + strings.Repeat("b", 61): true,  // 253 bytes
		pre + strings.Repeat("b", 62): false, // 254 bytes
	})
}
