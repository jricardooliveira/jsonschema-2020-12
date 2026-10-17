package main

import (
	"strings"
	"testing"
)

const metaCLI = "../../testdata/metaschema/cli/"

// Core 8.1.1 / 8.2.1: a schema failing its meta-schema, an unsupported
// '$schema' or a refused built-in address is exit 2, empty stdout and one
// value-free stderr line.
func TestCLIMetaSchemaErrors(t *testing.T) {
	doc := metaCLI + "document.json"
	cases := []struct {
		name      string
		args      []string
		want      []string // text stderr must hold
		forbidden []string
	}{
		{"invalid schema (core 8.1.1)", []string{"validate", metaCLI + "invalid_schema.json", doc},
			[]string{"invalid_schema.json", `"/enum"`, `rule "type"`}, []string{"secret-value-123"}},
		{"unsupported draft (core 8.1.1)", []string{"validate", metaCLI + "draft07_schema.json", doc},
			[]string{"unsupported $schema: draft-07; only 2020-12 is supported"}, nil},
		{"custom meta-schema failure (core 8.1.1)",
			[]string{"validate", "--ref", "https://x.test/meta=" + metaCLI + "custom_meta_schema.json", metaCLI + "custom_meta.json", doc},
			[]string{"custom_meta.json", `"required"`}, nil},
		{"built-in address mismatch (core 8.2.1)",
			[]string{"validate", "--ref", metaCLI + "builtin_address_mismatch.json", metaCLI + "valid_schema.json", doc},
			[]string{"https://json-schema.org/draft/2020-12/schema", "builtin_address_mismatch.json"}, nil},
	}
	for _, c := range cases {
		code, out, errOut := do(c.args, "")
		wantErr(t, c.name, code, out, errOut)
		for _, w := range c.want {
			if !strings.Contains(errOut, w) {
				t.Errorf("%s: stderr %q lacks %q", c.name, errOut, w)
			}
		}
		for _, f := range c.forbidden {
			if strings.Contains(errOut, f) {
				t.Errorf("%s: stderr %q must not contain %q", c.name, errOut, f)
			}
		}
	}
}

// Core 8.1.1: a valid schema gives a normal report.
func TestCLIMetaSchemaValid(t *testing.T) {
	code, out, errOut := do([]string{"validate", metaCLI + "valid_schema.json", metaCLI + "document.json"}, "")
	if code != exitValid || strings.Count(out, "\n") != 1 || errOut != "" {
		t.Errorf("code=%d out=%q err=%q", code, out, errOut)
	}
}
