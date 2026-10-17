package main

import (
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/eval"
)

// Broken references, ids and anchors are exit 2 whatever the document is.
func TestBrokenSchemasExitError(t *testing.T) {
	doc := write(t, "d.json", `1`)
	for name, schema := range map[string]string{
		"unreached unresolved ref": `{"properties":{"a":{"$ref":"#/nope"}}}`,
		"bad percent-encoding":     `{"$ref":"#/a%zz"}`,
		"duplicate anchor":         `{"allOf":[{"$anchor":"q"},{"$anchor":"q"}]}`,
		"pure ref cycle":           `{"$ref":"#"}`,
		"other resource":           `{"$ref":"other.json"}`,
		"non-empty $id fragment":   `{"$id":"http://x/y#f"}`,
	} {
		code, out, errOut := do([]string{"validate", write(t, "s.json", schema), doc}, "")
		wantErr(t, name, code, out, errOut)
	}
}

// core 8.2.3.1: a loop that never descends is an error at evaluation time.
func TestEvaluationLoopExitError(t *testing.T) {
	s := write(t, "s.json", `{"$ref":"#","type":"object"}`)
	code, out, errOut := do([]string{"validate", s, write(t, "d.json", `{}`)}, "")
	wantErr(t, "loop", code, out, errOut)
}

// Recursion that descends past the evaluation depth limit is exit 2, not a crash.
func TestEvaluationDepthExitError(t *testing.T) {
	k := 10
	schema := `{"items":` + strings.Repeat(`{"allOf":[`, k) + `{"$ref":"#"}` + strings.Repeat(`]}`, k) + `}`
	n := eval.MaxEvalDepth/(k+2) + 10
	doc := strings.Repeat("[", n) + strings.Repeat("]", n)
	code, out, errOut := do([]string{"validate", write(t, "s.json", schema), write(t, "d.json", doc)}, "")
	wantErr(t, "depth", code, out, errOut)
}

// core 9.1: --base-uri changes how a relative $ref resolves.
func TestBaseURIFlag(t *testing.T) {
	s := write(t, "s.json", relRef)
	d := write(t, "d.json", `1`)
	if code, _, e := do([]string{"validate", "--base-uri", "http://x/y/s.json", s, d}, ""); code != exitValid {
		t.Errorf("code=%d err=%q", code, e)
	}
	code, out, errOut := do([]string{"validate", "--base-uri", "http://x/z/s.json", s, d}, "")
	wantErr(t, "wrong base", code, out, errOut)
	code, out, errOut = do([]string{"validate", "--base-uri", "rel", s, d}, "")
	wantErr(t, "relative base", code, out, errOut)
	code, out, errOut = do([]string{"validate", s, d}, "")
	wantErr(t, "default base", code, out, errOut)
}

const relRef = `{"$defs":{"e":{"$id":"http://x/y/e.json"}},"$ref":"e.json"}`
