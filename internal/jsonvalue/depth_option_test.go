package jsonvalue

import (
	"errors"
	"strings"
	"testing"
)

func nest(n int) string { return strings.Repeat(`{"a":`, n) + "1" + strings.Repeat("}", n) }

// NumberLimits.MaxDepth caps nesting at, below and above the setting.
func TestDecodeDepthOption(t *testing.T) {
	const max = 4
	lim := NumberLimits{MaxDepth: max}
	for _, n := range []int{max - 1, max} {
		if _, err := DecodeLimited([]byte(nest(n)), lim); err != nil {
			t.Fatalf("depth %d: %v", n, err)
		}
	}
	var le *LimitError
	if _, err := DecodeLimited([]byte(nest(max+1)), lim); !errors.As(err, &le) || le.Limit != "nesting depth" || le.Max != max {
		t.Fatalf("over cap: %v", err)
	}
}

// Zero and negative fall back to MaxDepth; a value above encoding/json's own
// ceiling still ends in a LimitError, never a syntax error.
func TestDecodeDepthDefaultAndCeiling(t *testing.T) {
	for _, d := range []int{0, -1} {
		if got := (NumberLimits{MaxDepth: d}).resolved().MaxDepth; got != MaxDepth {
			t.Errorf("MaxDepth %d resolved to %d", d, got)
		}
	}
	deep := strings.Repeat("[", MaxDepth+1) + strings.Repeat("]", MaxDepth+1)
	var le *LimitError
	if _, err := DecodeLimited([]byte(deep), NumberLimits{MaxDepth: 2 * MaxDepth}); !errors.As(err, &le) || le.Limit != "nesting depth" {
		t.Fatalf("got %v", err)
	}
	if _, err := Decode([]byte(deep[1 : len(deep)-1])); err != nil {
		t.Fatalf("default at cap: %v", err)
	}
}
