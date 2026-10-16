package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func lines(t *testing.T, out string) []map[string]json.RawMessage {
	t.Helper()
	if out == "" || !strings.HasSuffix(out, "\n") {
		t.Fatalf("bad stdout %q", out)
	}
	var res []map[string]json.RawMessage
	for _, l := range strings.Split(strings.TrimSuffix(out, "\n"), "\n") {
		var m map[string]json.RawMessage
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Fatalf("line %q: %v", l, err)
		}
		res = append(res, m)
	}
	return res
}

func TestMultiDocExitCodes(t *testing.T) {
	schema := write(t, "s.json", `{"type":"number"}`)
	good := write(t, "good.json", `1`)
	bad := write(t, "bad.json", `"x"`)
	broken := write(t, "broken.json", `{`)
	cases := []struct {
		name string
		docs []string
		code int
	}{
		{"valid valid", []string{good, good}, exitValid},
		{"valid invalid", []string{good, bad}, exitInvalid},
		{"valid broken", []string{good, broken}, exitError},
		{"invalid broken", []string{bad, broken}, exitError},
		{"broken first", []string{broken, bad, good}, exitError},
	}
	for _, c := range cases {
		code, out, errOut := do(append([]string{"validate", schema}, c.docs...), "")
		if code != c.code || errOut != "" {
			t.Errorf("%s: code=%d err=%q", c.name, code, errOut)
		}
		if got := len(lines(t, out)); got != len(c.docs) {
			t.Errorf("%s: %d lines, want %d", c.name, got, len(c.docs))
		}
	}
}

func TestMultiDocMixedLines(t *testing.T) {
	schema := write(t, "s.json", `{"type":"number"}`)
	good := write(t, "good.json", `1`)
	bad := write(t, "bad.json", `"`+secretValue+`"`)
	broken := write(t, "broken.json", `{"`+secretValue+`" !}`)
	missing := good + ".nope"
	code, out, _ := do([]string{"validate", schema, good, bad, broken, missing, "-"}, `2`)
	if code != exitError {
		t.Errorf("code=%d", code)
	}
	if strings.Contains(out, secretValue) {
		t.Errorf("leak: %q", out)
	}
	ls := lines(t, out)
	if len(ls) != 5 {
		t.Fatalf("lines=%d", len(ls))
	}
	wantDoc := []string{good, bad, broken, missing, "-"}
	for i, l := range ls {
		var d string
		if err := json.Unmarshal(l["document"], &d); err != nil || d != wantDoc[i] {
			t.Errorf("line %d document=%s", i, l["document"])
		}
	}
	// standard output objects sit unchanged inside "output"
	if string(ls[0]["output"]) != `{"valid":true}` || string(ls[4]["output"]) != `{"valid":true}` {
		t.Errorf("valid outputs: %s %s", ls[0]["output"], ls[4]["output"])
	}
	if !strings.HasPrefix(string(ls[1]["output"]), `{"valid":false,"errors":[`) {
		t.Errorf("invalid output: %s", ls[1]["output"])
	}
	for _, i := range []int{2, 3} {
		if _, has := ls[i]["output"]; has || len(ls[i]["error"]) == 0 {
			t.Errorf("line %d should carry error only: %v", i, ls[i])
		}
	}
}

func TestMultiDocFlagShape(t *testing.T) {
	schema := write(t, "s.json", `{"type":"number"}`)
	good := write(t, "good.json", `1`)
	bad := write(t, "bad.json", `"x"`)
	code, out, _ := do([]string{"validate", "-o", "flag", schema, good, bad}, "")
	ls := lines(t, out)
	if code != exitInvalid || string(ls[0]["output"]) != `{"valid":true}` || string(ls[1]["output"]) != `{"valid":false}` {
		t.Errorf("code=%d out=%q", code, out)
	}
}

func TestSchemaErrorWithSeveralDocuments(t *testing.T) {
	doc := write(t, "d.json", `1`)
	cases := map[string]string{
		"broken json": `{`,
		"not schema":  `5`,
		"bad type":    `{"type":"bogus"}`,
	}
	for name, s := range cases {
		schema := write(t, "s.json", s)
		code, out, errOut := do([]string{"validate", schema, doc, doc, "-"}, `1`)
		wantErr(t, name, code, out, errOut)
	}
}

func TestSingleBrokenDocumentStdoutEmpty(t *testing.T) {
	schema := write(t, "s.json", `true`)
	broken := write(t, "b.json", `{`)
	code, out, errOut := do([]string{"validate", schema, broken}, "")
	wantErr(t, "single broken", code, out, errOut)
}

func TestStdinTwiceDoesNotHang(t *testing.T) {
	schema := write(t, "s.json", `true`)
	code, out, _ := do([]string{"validate", schema, "-", "-"}, `1`)
	ls := lines(t, out)
	if code != exitError || len(ls) != 2 || string(ls[0]["output"]) != `{"valid":true}` || len(ls[1]["error"]) == 0 {
		t.Errorf("code=%d out=%q", code, out)
	}
}
