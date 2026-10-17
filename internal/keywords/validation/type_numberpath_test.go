package validation_test

// Number paths for type "integer": JSON Schema validation spec section 6.1.1
// (1.0 is an integer). Fast path = mantissa fits int64; big path = it does not.

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/neoface/jsonschema"
)

// npValid compiles schema, validates doc and reports validity.
func npValid(t *testing.T, schema, doc string) bool {
	t.Helper()
	s, err := jsonschema.Compile([]byte(schema))
	if err != nil {
		t.Fatalf("%.80s: %v", schema, err)
	}
	r, err := s.Validate([]byte(doc))
	if err != nil {
		t.Fatalf("%.80s: %v", doc, err)
	}
	return r.Valid()
}

// Numbers at the default digit and exponent caps (read from the code).
// Fast = mantissa fits int64, big = it does not.
var (
	maxExp     = jsonschema.DefaultMaxNumberExponent
	capExp     = strconv.Itoa(maxExp)                                   // exponent exactly at the cap
	capExpM1   = strconv.Itoa(maxExp - 1)                               // one below, for trailing-zero forms
	fastAtExp  = "9223372036854775807e" + capExp                        // int64 max mantissa at the cap
	fastAtExp0 = "92233720368547758070e" + capExpM1                     // same value, trailing zero (demoted)
	bigAtExp   = "9223372036854775808e" + capExp                        // int64 max+1 mantissa at the cap
	tinyAtExp  = "1e-" + capExp                                         // smallest exponent at the cap
	nines      = strings.Repeat("9", jsonschema.DefaultMaxNumberDigits) // digits at the cap (big)
)

// npLimit checks that every doc past a number cap is a limit error, never
// a valid or invalid answer, under schema.
func npLimit(t *testing.T, schema string) {
	t.Helper()
	s, err := jsonschema.Compile([]byte(schema))
	if err != nil {
		t.Fatal(err)
	}
	for doc, want := range map[string]jsonschema.Limit{
		"1e" + strconv.Itoa(maxExp+1):                  jsonschema.LimitNumberExponent,
		"1e-" + strconv.Itoa(maxExp+1):                 jsonschema.LimitNumberExponent,
		nines + "9":                                    jsonschema.LimitNumberDigits,
		"0." + strings.Repeat("0", len(nines)-1) + "1": jsonschema.LimitNumberDigits,
	} {
		_, err := s.Validate([]byte(doc))
		var le *jsonschema.LimitError
		if !errors.As(err, &le) || le.Limit != want {
			t.Errorf("%s past cap (%.20s...): got %v, want %s", schema, doc, err, want)
		}
	}
}

func TestTypeIntegerCaps(t *testing.T) {
	for doc, want := range map[string]bool{
		"1e" + capExp: true, tinyAtExp: false, fastAtExp: true, fastAtExp0: true, bigAtExp: true,
		nines: true, "9223372036854775807e-" + capExp: false, nines + "e-1": false,
	} {
		if got := npValid(t, `{"type":"integer"}`, doc); got != want {
			t.Errorf("integer %.30s: got %v want %v", doc, got, want)
		}
	}
	npLimit(t, `{"type":"integer"}`)
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
