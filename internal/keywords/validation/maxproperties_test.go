package validation_test

import (
	"strings"
	"testing"
)

// The testdata/validation/maxproperties.json cases run through the root TestSpecCases.

func TestMaxPropertiesMessageHasNoKeys(t *testing.T) {
	out := basicOut(t, `{"maxProperties":1}`, `{"alice@example.com":1,"b":2}`)
	if !strings.Contains(out, "maxProperties") || strings.Contains(out, "alice") {
		t.Errorf("bad output: %s", out)
	}
}
