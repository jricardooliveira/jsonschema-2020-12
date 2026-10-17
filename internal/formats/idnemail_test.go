package formats

import "testing"

// Basic idn-email checks (see idnemail.go for the gaps).
func TestIDNEmail(t *testing.T) {
	table(t, "idn-email", map[string]bool{
		"user@example.com": true, `"john doe"@example.com`: true,
		"用户@例え.jp": true, "üser@example.com": true, "üser@[192.168.0.1]": true,
		"": false, "üser": false, "@example.com": false, "üser@": false,
		"ü ser@example.com": false, "ü..ser@example.com": false,
		"üser@-a.com": false, "üser@example.com.": false, "user@example.com.": false,
	})
}
