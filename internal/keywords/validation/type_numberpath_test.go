package validation_test

// Number paths for type "integer": JSON Schema validation spec section 6.1.1
// (1.0 is an integer). Fast path = mantissa fits int64; big path = it does not.

import (
	"testing"

	"github.com/neoface/jsonschema"
)

// npValid compiles schema, validates doc and reports validity.
func npValid(t *testing.T, schema, doc string) bool {
	t.Helper()
	s, err := jsonschema.Compile([]byte(schema))
	if err != nil {
		t.Fatalf("%s: %v", schema, err)
	}
	r, err := s.Validate([]byte(doc))
	if err != nil {
		t.Fatalf("%s: %v", doc, err)
	}
	return r.Valid()
}

func TestTypeIntegerNumberPaths(t *testing.T) {
	cases := []struct {
		doc  string
		want bool
	}{
		{"1", true}, {"1.0", true}, {"1e0", true}, {"100", true}, {"1e2", true},
		{"-0.0", true}, {"0", true}, {"1.5", false}, {"0.1", false}, {"1e-1", false},
		{"9223372036854775807", true},       // int64 max (fast)
		{"9223372036854775807.0", true},     // fast, trailing zero folds
		{"9223372036854775808", true},       // max+1 (big)
		{"9223372036854775808.0", true},     // big
		{"12345678901234567890123.0", true}, // big
		{"1.00000000000000000000001", false},
		{"1200000000000000000000e-5", true},
		{"12e-1", false},
	}
	for _, c := range cases {
		if got := npValid(t, `{"type":"integer"}`, c.doc); got != c.want {
			t.Errorf("integer %s: got %v want %v", c.doc, got, c.want)
		}
		if !npValid(t, `{"type":"number"}`, c.doc) {
			t.Errorf("number %s: want valid", c.doc)
		}
	}
}
