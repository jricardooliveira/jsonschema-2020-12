package formats

import (
	"strings"
	"testing"
)

// Basic idn-hostname checks (see idnhostname.go for the gaps).
func TestIDNHostname(t *testing.T) {
	table(t, "idn-hostname", map[string]bool{
		"example.com": true, "bücher.example": true, "例え.jp": true,
		"example.com.": false, "-a.com": false, "a-.com": false, "": false,
		"bü cher.de": false, "bü\tcher.de": false, "bü!cher.de": false, "bü..de": false,
	})
	a, _ := Lookup("hostname")
	b, _ := Lookup("idn-hostname")
	for _, h := range []string{"a_b.com", "a b", "a..b", "ok.example", "x-.y"} {
		if a(h) != b(h) {
			t.Errorf("ASCII %q differs from hostname", h)
		}
	}
}

// Non-ASCII lengths count code points: 63 per label, 253 overall.
func TestIDNHostnameLengths(t *testing.T) {
	l := "ü" + strings.Repeat("a", 62) + "." // 63 code points + dot
	pre := l + l + l                         // 192 code points, over 192 bytes
	table(t, "idn-hostname", map[string]bool{
		strings.Repeat("ü", 63) + ".de": true,
		strings.Repeat("ü", 64) + ".de": false,
		pre + strings.Repeat("b", 61):   true,  // 253 code points
		pre + strings.Repeat("b", 62):   false, // 254 code points
	})
}
