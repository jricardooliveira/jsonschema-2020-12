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

func TestMaxAnnotationsFlag(t *testing.T) {
	schema := write(t, "s.json", `{"title":"T","description":"D"}`)
	code, out, errOut := do([]string{"validate", "-max-annotations", "1", schema}, `1`)
	if code != exitError || out != "" || !strings.Contains(errOut, "annotations") {
		t.Fatalf("limit exceeded: code=%d out=%q err=%q", code, out, errOut)
	}
	if code, _, _ := do([]string{"validate", "-max-annotations", "2", schema}, `1`); code != exitValid {
		t.Fatalf("at limit: code=%d", code)
	}
	for _, v := range []string{"0", "-1", "x"} {
		code, out, errOut := do([]string{"validate", "-max-annotations", v, schema}, `1`)
		if code != exitError || out != "" || errOut == "" {
			t.Errorf("value %s: code=%d out=%q err=%q", v, code, out, errOut)
		}
	}
}
