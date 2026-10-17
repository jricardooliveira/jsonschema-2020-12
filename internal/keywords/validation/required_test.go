package validation_test

import (
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12"
)

// The testdata/validation/required.json cases run through the root TestSpecCases.

func basicOut(t *testing.T, schema, doc string) string {
	t.Helper()
	s, err := jsonschema.Compile([]byte(schema))
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Validate([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	out, err := r.Output("basic")
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestRequiredReportsMissingInSchemaOrder(t *testing.T) {
	out := basicOut(t, `{"required":["zed","email","alpha"]}`, `{"alpha":"secret"}`)
	z, e := strings.Index(out, "missing required property: zed"), strings.Index(out, "missing required property: email")
	if z < 0 || e < 0 || z > e {
		t.Errorf("order/message wrong: %s", out)
	}
	if strings.Contains(out, "missing required property: alpha") || strings.Contains(out, "secret") {
		t.Errorf("bad output: %s", out)
	}
}

func TestAllFailuresInOrderAndDeterministic(t *testing.T) {
	s, d := `{"properties":{"b":{"type":"string"},"a":{"type":"integer"}},"required":["zed","email"]}`, `{"b":1,"a":"x"}`
	out := basicOut(t, s, d)
	a, b := strings.Index(out, `"instanceLocation":"/a"`), strings.Index(out, `"instanceLocation":"/b"`)
	z, e := strings.Index(out, "missing required property: zed"), strings.Index(out, "missing required property: email")
	if a < 0 || b < 0 || z < 0 || e < 0 || !(a < b && b < z && z < e) {
		t.Errorf("want /a, /b, zed, email in order: %s", out)
	}
	if out != basicOut(t, s, d) {
		t.Error("output differs between runs")
	}
}

func TestRequiredSchemaErrors(t *testing.T) {
	for _, s := range []string{`{"required":"a"}`, `{"required":[1]}`, `{"dependentRequired":[]}`, `{"dependentRequired":{"a":"b"}}`, `{"dependentRequired":{"a":[1]}}`, `{"maxProperties":"x"}`, `{"minProperties":null}`} {
		if _, err := jsonschema.Compile([]byte(s)); err == nil {
			t.Errorf("%s: want schema error", s)
		}
	}
}
