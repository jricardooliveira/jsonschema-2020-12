package formats

import (
	"strings"
	"testing"
)

// RFC 5321 4.1.2, 4.1.3.
func TestEmail(t *testing.T) {
	table(t, "email", map[string]bool{
		"user@example.com": true, "a.b+c@example.com": true,
		`"john doe"@example.com`: true, `"a\"b@c"@example.com`: true,
		"user@[192.168.0.1]": true, "user@[IPv6:::1]": true, "user@[IPv6:2001:db8::1]": true,
		strings.Repeat("a", 64) + "@example.com": true,
		"":                                       false, "user": false, "@example.com": false, "user@": false,
		"user@example.com.": false, ".a@example.com": false, "a..b@example.com": false,
		"a b@example.com": false, `"unterminated@example.com`: false,
		"user@[192.168.01.1]": false, "user@[::1]": false, "user@[IPv6:zz]": false,
		"user@-a.com":                            false,
		strings.Repeat("a", 65) + "@example.com": false,
		"a@" + strings.Repeat("a.", 126) + "com": false,
		// RFC 5321 2.3.11 / 4.1.2: address-literal tag is case-insensitive.
		"user@[ipv6:::1]": true, "user@[IPV6:::1]": true,
	})
	// RFC 5321 4.5.3.1: only the 254-byte whole-address cap fails here; the
	// local part (64) and domain (190, under 253) are each within limits.
	local := strings.Repeat("a", 64)
	dom := strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "."
	table(t, "email", map[string]bool{
		local + "@" + dom + strings.Repeat("d", 61): true,  // 254 bytes
		local + "@" + dom + strings.Repeat("d", 62): false, // 255 bytes
	})
}
