package validation_test

import (
	"strings"
	"testing"

	"github.com/neoface/jsonschema"
)

// The testdata/validation/maxItems.json cases run through the root TestSpecCases.

func TestItemCountSchemaErrors(t *testing.T) {
	for _, s := range []string{`{"maxItems":"x"}`, `{"minItems":"x"}`, `{"maxItems":null}`, `{"minItems":[1]}`} {
		if _, err := jsonschema.Compile([]byte(s)); err == nil {
			t.Errorf("%s: want schema error", s)
		}
	}
}

func TestItemCountErrorHasNoValues(t *testing.T) {
	s, err := jsonschema.Compile([]byte(`{"maxItems":1}`))
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Validate([]byte(`["secret-a","secret-b"]`))
	if err != nil || r.Valid() {
		t.Fatalf("want invalid, got %v %v", r.Valid(), err)
	}
	out, err := r.Output("basic")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "maxItems") || strings.Contains(string(out), "secret") {
		t.Errorf("bad output: %s", out)
	}
}
