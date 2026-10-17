package main

import "testing"

// A custom meta-schema that closes the 2020-12 one (unevaluatedProperties
// false over an allOf $ref to it): an unknown keyword in a schema is a schema
// error, exit 2; a schema with listed keywords only validates.
func TestCustomStrictMetaSchemaCLI(t *testing.T) {
	meta := write(t, "strict-meta.json", `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://example.com/strict-meta",
  "$dynamicAnchor": "meta",
  "allOf": [{"$ref": "https://json-schema.org/draft/2020-12/schema"}],
  "unevaluatedProperties": false
}`)
	doc := write(t, "doc.json", `{"a":1}`)
	const hdr = `"$schema":"https://example.com/strict-meta",`
	for _, c := range []struct {
		name, schema string
		want         int
	}{
		{"unknown at root", `{` + hdr + `"bogus":1}`, exitError},
		{"unknown nested", `{` + hdr + `"properties":{"a":{"bogus":1}}}`, exitError},
		{"listed only", `{` + hdr + `"properties":{"a":{"type":"integer"}}}`, exitValid},
	} {
		sch := write(t, "schema.json", c.schema)
		if code, _, stderr := do([]string{"validate", "--ref", meta, sch, doc}, ""); code != c.want {
			t.Errorf("%s: exit %d, want %d (stderr %q)", c.name, code, c.want, stderr)
		}
	}
}
