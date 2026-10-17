package validation_test

import (
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12"
)

// The testdata/validation/minContains.json cases run through the root TestSpecCases.

func TestContainsBoundsSchemaErrors(t *testing.T) {
	for _, s := range []string{`{"minContains":"x"}`, `{"maxContains":[1]}`} {
		if _, err := jsonschema.Compile([]byte(s)); err == nil {
			t.Errorf("%s: want schema error", s)
		}
	}
}
