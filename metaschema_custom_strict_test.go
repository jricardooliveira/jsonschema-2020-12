package jsonschema

import (
	"os"
	"path/filepath"
	"testing"
)

// strictMeta is a custom meta-schema that closes the 2020-12 one:
// unevaluatedProperties false over an allOf $ref to it (core 8.1.1, 10.3.2).
// The $dynamicAnchor makes the nested "properties" subschemas strict too.
const strictMeta = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://example.com/strict-meta",
  "$dynamicAnchor": "meta",
  "allOf": [{"$ref": "https://json-schema.org/draft/2020-12/schema"}],
  "unevaluatedProperties": false
}`

func strictRegistry(t *testing.T) Registry {
	t.Helper()
	p := filepath.Join(t.TempDir(), "strict-meta.json")
	if err := os.WriteFile(p, []byte(strictMeta), 0o600); err != nil {
		t.Fatal(err)
	}
	return Registry{Files: []RegistryFile{{Path: p}}}
}

func TestCustomStrictMetaSchema(t *testing.T) {
	reg := strictRegistry(t)
	const hdr = `"$schema":"https://example.com/strict-meta",`
	for name, src := range map[string]string{
		"unknown at root":   `{` + hdr + `"bogus":1}`,
		"unknown in nested": `{` + hdr + `"properties":{"a":{"bogus":1}}}`,
	} {
		if _, err := CompileWithRegistry([]byte(src), Options{}, reg); err == nil {
			t.Errorf("%s: loaded, want an invalid-schema error", name)
		}
	}
	s, err := CompileWithRegistry([]byte(`{`+hdr+`"type":"object","properties":{"a":{"type":"integer"}}}`), Options{}, reg)
	if err != nil {
		t.Fatalf("listed keywords only: %v", err)
	}
	if !mustValidate(t, s, `{"a":1}`) || mustValidate(t, s, `{"a":"x"}`) {
		t.Fatal("wrong verdicts")
	}
}
