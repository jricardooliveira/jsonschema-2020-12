package validation_test

import (
	"strings"
	"testing"

	"github.com/neoface/jsonschema"
)

// The testdata/validation/maxContains.json cases run through the root TestSpecCases.

func TestMaxContainsErrorHasNoValues(t *testing.T) {
	s, err := jsonschema.Compile([]byte(`{"contains":{"type":"string"},"maxContains":1}`))
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Validate([]byte(`["secret-a","secret-b"]`))
	if err != nil || r.Valid() {
		t.Fatalf("valid=%v err=%v", r.Valid(), err)
	}
	out, _ := r.Output("basic")
	if !strings.Contains(string(out), "maxContains") || strings.Contains(string(out), "secret") {
		t.Errorf("bad output: %s", out)
	}
}
