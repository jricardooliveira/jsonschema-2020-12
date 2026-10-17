package formats

import "testing"

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
