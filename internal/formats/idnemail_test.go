package formats

import (
	"strings"
	"testing"
)

// Basic idn-email checks (see idnemail.go for the gaps).
func TestIDNEmail(t *testing.T) {
	table(t, "idn-email", map[string]bool{
		"user@example.com": true, `"john doe"@example.com`: true,
		"用户@例え.jp": true, "üser@example.com": true, "üser@[192.168.0.1]": true,
		"ü.s+e_r@example.com": true,
		"":                    false, "üser": false, "@example.com": false, "üser@": false,
		"ü ser@example.com": false, "ü..ser@example.com": false,
		"üser@-a.com": false, "üser@example.com.": false, "user@example.com.": false,
	})
}

// A non-ASCII unquoted local part allows only atext and '.' among ASCII
// characters (RFC 6531 3.3 extends RFC 5322 atext).
func TestIDNEmailStrayPunctuation(t *testing.T) {
	cases := map[string]bool{}
	for _, c := range []string{",", "(", ")", "<", ">", ";", ":", "[", "]", `\`, `"`} {
		cases["ü"+c+"ser@example.com"] = false
	}
	table(t, "idn-email", cases)
}

// Non-ASCII length limits: local part 64 bytes, whole address 254 bytes.
func TestIDNEmailLengths(t *testing.T) {
	local := "ü" + strings.Repeat("a", 62) // 64 bytes
	dom := strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "."
	table(t, "idn-email", map[string]bool{
		local + "@example.com":                      true,
		local + "a@example.com":                     false, // 65-byte local part
		local + "@" + dom + strings.Repeat("d", 61): true,  // 254 bytes
		local + "@" + dom + strings.Repeat("d", 62): false, // 255 bytes
	})
}
