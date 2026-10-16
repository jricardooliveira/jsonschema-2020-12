package uri

import (
	"reflect"
	"testing"
)

func TestParsePointer(t *testing.T) {
	for in, want := range map[string][]string{
		"":             nil,
		"/$defs/a~1b":  {"$defs", "a/b"},
		"/$defs/a~0b":  {"$defs", "a~b"},
		"/$defs/a%25b": {"$defs", "a%b"},
		"/a%7e1b":      {"a/b"}, // decode first: %7e is '~', then ~1
		"/a%2Fb":       {"a/b"},
		"/a%257E1b":    {"a%7E1b"}, // %25 decodes once only
		"//":           {"", ""},
	} {
		got, err := ParsePointer(in)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Errorf("%q: %q,%v want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"/a%zz", "a/b", "/a~2", "/a~"} {
		if _, err := ParsePointer(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}
