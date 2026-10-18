package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12"
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

// A detailed or verbose report over DefaultMaxReportBytes is exit 2 with
// empty stdout and one stderr line; with several documents that document
// gets the {document, error} line and the worst result wins. Each item
// annotates a 1 MiB title, so the report passes the cap by one item.
func TestNestedReportOverCapCLI(t *testing.T) {
	const mib = 1 << 20
	schema := write(t, "s.json", `{"items":{"title":"`+strings.Repeat("t", mib)+`"}}`)
	big := write(t, "big.json", "["+strings.TrimSuffix(strings.Repeat("0,", jsonschema.DefaultMaxReportBytes/mib+1), ",")+"]")
	good := write(t, "g.json", `{}`)
	for _, shape := range []string{"detailed", "verbose"} {
		code, out, errOut := do([]string{"validate", "-output", shape, schema, big}, "")
		wantErr(t, shape, code, out, errOut)
		if !strings.Contains(errOut, "report bytes") {
			t.Errorf("%s: stderr %q", shape, errOut)
		}
		code, out, _ = do([]string{"validate", "-output", shape, schema, good, big}, "")
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		if code != exitError || len(lines) != 2 {
			t.Fatalf("%s multi: code=%d lines=%d", shape, code, len(lines))
		}
		var line struct{ Document, Error string }
		if err := json.Unmarshal([]byte(lines[1]), &line); err != nil || line.Document != big || !strings.Contains(line.Error, "report bytes") {
			t.Errorf("%s: error line %q", shape, lines[1])
		}
	}
}

// Flag and basic CLI output and exit codes for a large failing document are
// the library output pinned by output_unchanged_test.go at the root.
func TestFlagBasicCLIUnchanged(t *testing.T) {
	const schemaText = `{"items":{"properties":{"a":{"type":"string"},"b":{"minimum":0}},"required":["c"]}}`
	docText := "[" + strings.TrimSuffix(strings.Repeat(`{"a":1,"b":-1},`, 3000), ",") + "]"
	schema := write(t, "s.json", schemaText)
	doc := write(t, "d.json", docText)
	s, err := jsonschema.Compile([]byte(schemaText))
	if err != nil {
		t.Fatal(err)
	}
	res, err := s.Validate([]byte(docText))
	if err != nil {
		t.Fatal(err)
	}
	for _, shape := range []string{"flag", "basic"} {
		want, err := res.Output(shape)
		if err != nil {
			t.Fatal(err)
		}
		code, out, _ := do([]string{"validate", "-output", shape, schema, doc}, "")
		if code != exitInvalid || out != string(want)+"\n" {
			t.Errorf("%s: code=%d, stdout differs from library output", shape, code)
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
