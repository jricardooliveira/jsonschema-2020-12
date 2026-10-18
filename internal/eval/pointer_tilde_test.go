package eval

import (
	"slices"
	"testing"
)

// The decoder itself is tested in internal/output/pointer_tilde_test.go; this
// checks that both pointer splitters route through it.
func TestTokensDecodeTildeSegments(t *testing.T) {
	p := "/a~1b/~01/plain/~0"
	want := []string{"a/b", "~1", "plain", "~"}
	if got := tokens(p); !slices.Equal(got, want) {
		t.Errorf("tokens(%q) = %q, want %q", p, got, want)
	}
	if got := appendTokens(nil, p); !slices.Equal(got, want) {
		t.Errorf("appendTokens(%q) = %q, want %q", p, got, want)
	}
}
