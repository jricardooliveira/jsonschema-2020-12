package compile

import (
	"net/url"
	"testing"
)

func TestEncodeFragment(t *testing.T) {
	cases := map[string]string{
		"/$defs/a b/type": "/$defs/a%20b/type",
		"/$defs/50%/type": "/$defs/50%25/type",
		"/$defs/é/type":   "/$defs/%C3%A9/type",
		"/a~1b/~0c":       "/a~1b/~0c",
		"":                "",
	}
	for in, want := range cases {
		got := encodeFragment(in)
		if got != want {
			t.Errorf("%q: got %q want %q", in, got, want)
		}
		u, err := url.Parse("http://x/r#" + got)
		if err != nil || u.Fragment != in {
			t.Errorf("%q: parse err=%v fragment=%q", in, err, u.Fragment)
		}
	}
}

func TestAbsoluteLocationEncodes(t *testing.T) {
	r := &Resource{URI: "http://x/r"}
	if got := r.AbsoluteLocation("/$defs/a b"); got != "http://x/r#/$defs/a%20b" {
		t.Errorf("got %q", got)
	}
}
