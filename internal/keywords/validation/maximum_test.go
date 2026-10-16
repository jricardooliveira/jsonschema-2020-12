package validation_test

// Derived from JSON Schema validation spec sections 6.2.2 (maximum) and 6.2.4 (minimum).

import (
	"testing"

	"github.com/neoface/jsonschema"
)

func TestMaximumBounds(t *testing.T) {
	cases := []struct {
		schema, doc string
		want        bool
	}{
		{`{"maximum":1}`, `1.0`, true},
		{`{"maximum":1}`, `1.0000000000000000000001`, false},
		{`{"maximum":12345678901234567890}`, `12345678901234567891`, false},
		{`{"maximum":12345678901234567890}`, `12345678901234567890`, true},
		{`{"maximum":0.30000000000000000001}`, `0.30000000000000000002`, false},
		{`{"maximum":1}`, `"5"`, true},
		{`{"maximum":1}`, `null`, true},
		{`{"maximum":1e10000}`, `1e-10000`, true},
	}
	runBounds(t, cases)
}

// Exponents far past the default cap: Cmp must decide by the adjusted
// exponent without expanding a power of ten.
func TestBoundsHostileScale(t *testing.T) {
	e := "999999999"
	o := jsonschema.Options{MaxNumberExponent: 2_000_000_000}
	cases := []struct {
		schema, doc string
		want        bool
	}{
		{`{"maximum":1}`, `1e` + e, false},
		{`{"maximum":1}`, `1e-` + e, true},
		{`{"minimum":1}`, `1e-` + e, false},
		{`{"exclusiveMaximum":1e` + e + `}`, `1e` + e, false},
		{`{"exclusiveMinimum":-1e` + e + `}`, `-1e-` + e, true},
		{`{"maximum":1e-` + e + `}`, `1.0e-` + e, true},
	}
	for _, c := range cases {
		if got := multipleOfValid(t, c.schema, c.doc, o); got != c.want {
			t.Errorf("%s vs %s: got %v", c.schema, c.doc, got)
		}
	}
}

func runBounds(t *testing.T, cases []struct {
	schema, doc string
	want        bool
}) {
	t.Helper()
	for _, c := range cases {
		s, err := jsonschema.Compile([]byte(c.schema))
		if err != nil {
			t.Fatal(err)
		}
		if r, err := s.Validate([]byte(c.doc)); err != nil || r.Valid() != c.want {
			t.Errorf("%s vs %s: got %v err %v want %v", c.schema, c.doc, r.Valid(), err, c.want)
		}
	}
}

func TestMaximumMinimumSchemaErrors(t *testing.T) {
	for _, s := range []string{`{"maximum":"1"}`, `{"minimum":null}`} {
		if _, err := jsonschema.Compile([]byte(s)); err == nil {
			t.Errorf("%s: want schema error", s)
		}
	}
}
