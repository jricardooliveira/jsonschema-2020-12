package core_test

import (
	"strings"
	"testing"

	jsonschema "github.com/jricardooliveira/jsonschema-2020-12"
)

// core 8.1.1, 8.1.2, 8.3: $schema, $vocabulary and $comment have no validation
// effect and produce no annotation.
func TestSilentCoreKeywords(t *testing.T) {
	for _, schema := range []string{
		`{"$schema":"https://json-schema.org/draft/2020-12/schema","type":"string"}`,
		`{"$comment":"c","type":"string"}`,
		`{"$vocabulary":{"https://json-schema.org/draft/2020-12/vocab/core":true},"type":"string"}`,
	} {
		s, err := jsonschema.Compile([]byte(schema))
		if err != nil {
			t.Fatalf("%s: %v", schema, err)
		}
		res, err := s.Validate([]byte(`"x"`))
		if err != nil {
			t.Fatal(err)
		}
		out, err := res.Output("basic")
		if err != nil {
			t.Fatal(err)
		}
		for _, k := range []string{"/$schema", "/$vocabulary", "/$comment"} {
			if strings.Contains(string(out), k) {
				t.Errorf("%s: report has annotation %s: %s", schema, k, out)
			}
		}
	}
}
