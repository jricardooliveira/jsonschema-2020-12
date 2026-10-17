package validation_test

import (
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12"
)

func TestMinLengthBoundary(t *testing.T) {
	s, err := jsonschema.Compile([]byte(`{"minLength":2}`))
	if err != nil {
		t.Fatal(err)
	}
	for doc, want := range map[string]bool{`"😀"`: false, `"😀😀"`: true, `5`: true} {
		if r, err := s.Validate([]byte(doc)); err != nil || r.Valid() != want {
			t.Errorf("%s: got %v %v", doc, r.Valid(), err)
		}
	}
}
