package validation_test

import (
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12"
)

// The testdata/validation/minItems.json cases run through the root TestSpecCases.

func TestMinItemsConcurrent(t *testing.T) {
	s, err := jsonschema.Compile([]byte(`{"minItems":2.0}`))
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan bool)
	for i := 0; i < 4; i++ {
		go func() {
			r, err := s.Validate([]byte(`[1]`))
			done <- err == nil && !r.Valid()
		}()
	}
	for i := 0; i < 4; i++ {
		if !<-done {
			t.Error("want invalid")
		}
	}
}
