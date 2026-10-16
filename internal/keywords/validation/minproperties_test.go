package validation_test

import (
	"strings"
	"testing"
)

// The testdata/validation/minproperties.json cases run through the root TestSpecCases.

func TestMinPropertiesFailureNamesKeyword(t *testing.T) {
	out := basicOut(t, `{"minProperties":2}`, `{"a":1}`)
	if !strings.Contains(out, "minProperties") {
		t.Errorf("bad output: %s", out)
	}
}
