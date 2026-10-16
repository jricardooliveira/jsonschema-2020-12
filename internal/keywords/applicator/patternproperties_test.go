package applicator_test

import (
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/keywords/validation"
)

// The testdata/applicator/patternproperties.json cases run through the root TestSpecCases.

func TestPatternPropertiesSchemaErrors(t *testing.T) {
	big := strings.Repeat("a", validation.DefaultMaxPatternSize+1)
	for _, s := range []string{
		`{"patternProperties":{"(?=a)":true}}`,
		`{"patternProperties":{"(":true}}`,
		`{"patternProperties":{"` + big + `":true}}`,
		`{"patternProperties":[]}`,
		`{"patternProperties":{"a":"x"}}`,
		`{"additionalProperties":true,"patternProperties":{"(?=a)":true}}`,
	} {
		if _, err := compile.Compile(mustDecode(t, s)); err == nil {
			t.Errorf("%.60s: want schema error", s)
		}
	}
}
