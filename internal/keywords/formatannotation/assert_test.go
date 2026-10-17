package formatannotation_test

import (
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/formats"
)

func init() { formats.Register("stub-only-x", func(s string) bool { return s == "x" }) }

func validate(t *testing.T, schema string, assert bool, doc string) bool {
	t.Helper()
	s, err := jsonschema.CompileWithOptions([]byte(schema), jsonschema.Options{FormatAssertion: assert})
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Validate([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	return r.Valid()
}

// validation 7.2: assertion only in the mode; unknown formats and non-strings pass.
func TestFormatAssertionMode(t *testing.T) {
	const stub = `{"format":"stub-only-x"}`
	cases := []struct {
		name, schema, doc string
		assert, valid     bool
	}{
		{"fails in mode", stub, `"y"`, true, false},
		{"passes in mode", stub, `"x"`, true, true},
		{"annotation only without mode", stub, `"y"`, false, true},
		{"unknown format passes", `{"format":"no-such"}`, `"y"`, true, true},
		{"non-string passes", stub, `5`, true, true},
	}
	for _, c := range cases {
		if got := validate(t, c.schema, c.assert, c.doc); got != c.valid {
			t.Errorf("%s: valid = %v", c.name, got)
		}
	}
}

// The failure report names locations, never the value.
func TestFormatAssertionReportHasNoValue(t *testing.T) {
	s, _ := jsonschema.CompileWithOptions([]byte(`{"format":"stub-only-x"}`), jsonschema.Options{FormatAssertion: true})
	r, _ := s.Validate([]byte(`"secret-value"`))
	out, err := r.Output("basic")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "secret-value") || !strings.Contains(string(out), "/format") {
		t.Errorf("output %s", out)
	}
}

// core 8.1.1: the meta-schema check never asserts format ($id is a uri-reference).
func TestMetaCheckNeverAssertsFormat(t *testing.T) {
	for _, assert := range []bool{false, true} {
		if _, err := jsonschema.CompileWithOptions([]byte(`{"$id":"http://example.com/a b"}`), jsonschema.Options{FormatAssertion: assert}); err != nil {
			t.Errorf("assert=%v: %v", assert, err)
		}
	}
}
