package formats

import "testing"

// RFC 3986 section 3 and RFC 3987 section 2.2.
func TestRFC3986Grammar(t *testing.T) {
	cases := []struct {
		s           string
		iri, scheme bool
		want        bool
	}{
		{"http://a/b?c#d", false, true, true},
		{"mailto:a@b", false, true, true},
		{"http://[::1]:80/", false, true, true},
		{"http://[v1.x]/", false, true, true},
		{"http://u:p@h:8080", false, true, true},
		{"urn:a:b", false, true, true},
		{"/a/b", false, true, false},
		{"http://exa mple.com", false, false, false},
		{"http://a/%zz", false, false, false},
		{"http://a/%4", false, false, false},
		{"http://a/ü", false, false, false},
		{"http://a/ü", true, true, true},
		{"http://a/?", true, true, true},
		{"http://a/", true, true, false},
		{"a:b:c", false, false, true},
		{"a/b:c", false, false, true},
		{"./a:b", false, false, true},
		{"1a:b", false, false, false},
		{"http://[::1", false, false, false},
		{"http://h:80x/", false, false, false},
		{"\xff", true, false, false},
		{"", false, false, true},
		{"#frag", false, false, true},
		{"a#b#c", false, false, false},
	}
	for _, c := range cases {
		if got := validRFC3986(c.s, c.iri, c.scheme); got != c.want {
			t.Errorf("validRFC3986(%q, iri=%v, scheme=%v) = %v", c.s, c.iri, c.scheme, got)
		}
	}
}
