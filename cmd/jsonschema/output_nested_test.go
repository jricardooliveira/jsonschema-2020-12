package main

import (
	"encoding/json"
	"strings"
	"testing"
)

const nestedCLISchema = `{"$id":"http://x/r","$ref":"#/$defs/a%20b%25%C3%A9","$defs":{"a b%é":{"type":"number"}}}`

func TestNestedShapesCLI(t *testing.T) {
	schema := write(t, "s.json", nestedCLISchema)
	good := write(t, "g.json", `1`)
	bad := write(t, "b.json", `"SECRET-VALUE-42"`)
	for _, shape := range []string{"detailed", "verbose"} {
		code, out, _ := do([]string{"validate", "-output", shape, schema, good}, "")
		var m map[string]any
		if code != exitValid || json.Unmarshal([]byte(out), &m) != nil || m["valid"] != true {
			t.Errorf("%s valid: code=%d out=%q", shape, code, out)
		}
		code, out, _ = do([]string{"validate", "-output", shape, schema, bad}, "")
		if code != exitInvalid || json.Unmarshal([]byte(out), &m) != nil || m["valid"] != false || strings.Count(strings.TrimRight(out, "\n"), "\n") != 0 {
			t.Errorf("%s invalid: code=%d out=%q", shape, code, out)
		}
		if strings.Contains(out, "SECRET") || !strings.Contains(out, "a%20b%25%C3%A9") {
			t.Errorf("%s: leak or unencoded location: %s", shape, out)
		}
		// several documents: one line each, worst result wins
		code, out, _ = do([]string{"validate", "-output", shape, schema, good, bad}, "")
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		if code != exitInvalid || len(lines) != 2 {
			t.Fatalf("%s multi: code=%d out=%q", shape, code, out)
		}
		for _, l := range lines {
			if !json.Valid([]byte(l)) || strings.Contains(l, "SECRET") {
				t.Errorf("%s bad line %q", shape, l)
			}
		}
		// error (missing document) wins over invalid
		code, _, _ = do([]string{"validate", "-output", shape, schema, bad, "/nonexistent/doc.json"}, "")
		if code != exitError {
			t.Errorf("%s: code=%d want error", shape, code)
		}
	}
}

func TestUnknownShapeCLIIsError(t *testing.T) {
	schema := write(t, "s.json", `true`)
	doc := write(t, "d.json", `1`)
	code, out, _ := do([]string{"validate", "-output", "nope", schema, doc}, "")
	if code != exitError || out != "" {
		t.Errorf("code=%d out=%q", code, out)
	}
}
