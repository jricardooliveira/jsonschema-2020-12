package main

import (
	"strings"
	"testing"
)

const annSecret = "zq-secret-instance-value"

func TestPassingBasicListsAnnotations(t *testing.T) {
	schema := write(t, "s.json", `{"title":"T","properties":{"a":{"description":"D"}}}`)
	code, out, errOut := do([]string{"validate", schema}, `{"a":"`+annSecret+`"}`)
	want := `{"valid":true,"annotations":[` +
		`{"valid":true,"instanceLocation":"","keywordLocation":"/properties","annotation":["a"]},` +
		`{"valid":true,"instanceLocation":"","keywordLocation":"/title","annotation":"T"},` +
		`{"valid":true,"instanceLocation":"/a","keywordLocation":"/properties/a/description","annotation":"D"}]}` + "\n"
	if code != exitValid || out != want || errOut != "" {
		t.Fatalf("code=%d out=%q err=%q", code, out, errOut)
	}
	if strings.Contains(out, annSecret) {
		t.Fatal("instance value leaked")
	}
}

func TestFailingBasicAndFlagHaveNoAnnotations(t *testing.T) {
	schema := write(t, "s.json", `{"title":"T","type":"number"}`)
	for _, shape := range []string{"basic", "flag"} {
		code, out, _ := do([]string{"validate", "-o", shape, schema}, `"`+annSecret+`"`)
		if code != exitInvalid || strings.Contains(out, "annotation") || strings.Contains(out, annSecret) {
			t.Errorf("%s: code=%d out=%q", shape, code, out)
		}
	}
	code, out, _ := do([]string{"validate", "-o", "flag", write(t, "p.json", `{"title":"T"}`)}, `1`)
	if code != exitValid || out != `{"valid":true}`+"\n" {
		t.Errorf("flag passing: code=%d out=%q", code, out)
	}
}

// Distinctive string, number and key values from the document never appear
// in any shape, passing or failing. The only names reported are ones the
// schema declares (evaluated property names are allowed in locations).
func TestNoInstanceValuesInAnyShape(t *testing.T) {
	const num, key = "987654321.125", "zq-secret-key"
	schema := write(t, "s.json", `{"title":"T","default":3,"properties":{"a":{"description":"D"},"n":{"type":"string"}}}`)
	pass := `{"a":"` + annSecret + `","` + key + `":` + num + `}`
	fail := `{"a":"` + annSecret + `","n":` + num + `,"` + key + `":"` + annSecret + `"}`
	for _, shape := range []string{"basic", "flag"} {
		for doc, wantCode := range map[string]int{pass: exitValid, fail: exitInvalid} {
			code, out, errOut := do([]string{"validate", "-o", shape, schema}, doc)
			if code != wantCode {
				t.Errorf("%s: code=%d want %d", shape, code, wantCode)
			}
			for _, v := range []string{annSecret, num, "987654321", key} {
				if strings.Contains(out, v) || strings.Contains(errOut, v) {
					t.Errorf("%s: value %q leaked: %q %q", shape, v, out, errOut)
				}
			}
		}
	}
	// Schema numbers render as written, not as {}.
	if _, out, _ := do([]string{"validate", schema}, `1`); !strings.Contains(out, `"keywordLocation":"/default","annotation":3}`) {
		t.Errorf("numeric annotation: %q", out)
	}
}

func TestMaxAnnotationsFlag(t *testing.T) {
	// The cap also covers the load-time meta-schema check (core 8.1.1), so
	// it leaves room for that; the long document then goes over it.
	schema := write(t, "s.json", `{"items":{"title":"T"}}`)
	code, out, errOut := do([]string{"validate", "-max-annotations", "100", schema}, `[`+strings.Repeat(`1,`, 200)+`1]`)
	if code != exitError || out != "" || !strings.Contains(errOut, "annotations") {
		t.Fatalf("limit exceeded: code=%d out=%q err=%q", code, out, errOut)
	}
	if code, _, _ := do([]string{"validate", "-max-annotations", "100", schema}, `[1]`); code != exitValid {
		t.Fatalf("under limit: code=%d", code)
	}
	// A cap the schema's own meta-check exceeds is exit 2, never a result.
	if code, out, _ := do([]string{"validate", "-max-annotations", "1", schema}, `1`); code != exitError || out != "" {
		t.Fatalf("meta-check over limit: code=%d out=%q", code, out)
	}
	for _, v := range []string{"0", "-1", "x"} {
		code, out, errOut := do([]string{"validate", "-max-annotations", v, schema}, `1`)
		if code != exitError || out != "" || errOut == "" {
			t.Errorf("value %s: code=%d out=%q err=%q", v, code, out, errOut)
		}
	}
}
