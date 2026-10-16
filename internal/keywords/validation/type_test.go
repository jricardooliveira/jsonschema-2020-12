package validation_test

import (
	"testing"

	"github.com/neoface/jsonschema"
)

// The testdata/validation/type.json cases run through the root TestSpecCases.

func TestTypeSchemaErrors(t *testing.T) {
	for _, s := range []string{`{"type":"foo"}`, `{"type":[]}`, `{"type":["string","string"]}`, `{"type":1}`, `{"type":[1]}`} {
		if _, err := jsonschema.Compile([]byte(s)); err == nil {
			t.Errorf("%s: want schema error", s)
		}
	}
}

func TestTypeIntegerAndArrayForm(t *testing.T) {
	s, err := jsonschema.Compile([]byte(`{"type":["integer","null"]}`))
	if err != nil {
		t.Fatal(err)
	}
	for doc, want := range map[string]bool{`1.0`: true, `null`: true, `1.5`: false, `"x"`: false} {
		if r, err := s.Validate([]byte(doc)); err != nil || r.Valid() != want {
			t.Errorf("%s: got %v %v", doc, r.Valid(), err)
		}
	}
}
