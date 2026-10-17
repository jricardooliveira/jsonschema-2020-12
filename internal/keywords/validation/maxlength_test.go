package validation_test

import (
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12"
)

func TestLengthSchemaErrors(t *testing.T) {
	for _, kw := range []string{"maxLength", "minLength"} {
		for _, v := range []string{`-1`, `1.5`, `"2"`, `null`} {
			if _, err := jsonschema.Compile([]byte(`{"` + kw + `":` + v + `}`)); err == nil {
				t.Errorf("%s %s: want schema error", kw, v)
			}
		}
	}
}

func TestLengthMessagesOmitInstance(t *testing.T) {
	const secret = "s3cr3t-xyz"
	for kw, schema := range map[string]string{"maxLength": `{"maxLength":2}`, "minLength": `{"minLength":99}`} {
		s, err := jsonschema.Compile([]byte(schema))
		if err != nil {
			t.Fatal(err)
		}
		r, err := s.Validate([]byte(`"` + secret + `"`))
		if err != nil || r.Valid() {
			t.Fatalf("%s: want invalid, got %v %v", kw, r.Valid(), err)
		}
		out, err := r.Output("basic")
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(out), secret) {
			t.Errorf("%s: output contains instance: %s", kw, out)
		}
	}
}
