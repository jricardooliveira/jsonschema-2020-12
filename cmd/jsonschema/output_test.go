package main

import (
	"strings"
	"testing"
)

const secretValue = "s3cr3t-value"

func TestShapesAndDefault(t *testing.T) {
	schema := write(t, "s.json", `{"type":"number"}`)
	doc := write(t, "d.json", `"`+secretValue+`"`)
	basic := `{"valid":false,"errors":[{"instanceLocation":"","keywordLocation":"/type","error":"value must be of type number"}]}` + "\n"
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"default", []string{"validate", schema, doc}, basic},
		{"-o basic", []string{"validate", "-o", "basic", schema, doc}, basic},
		{"--output basic", []string{"validate", "--output", "basic", schema, doc}, basic},
		{"-o flag", []string{"validate", "-o", "flag", schema, doc}, `{"valid":false}` + "\n"},
		{"--output flag", []string{"validate", "--output", "flag", schema, doc}, `{"valid":false}` + "\n"},
	}
	for _, c := range cases {
		code, out, errOut := do(c.args, "")
		if code != exitInvalid || out != c.want || errOut != "" || strings.Contains(out, secretValue) {
			t.Errorf("%s: code=%d out=%q err=%q", c.name, code, out, errOut)
		}
	}
}

func TestValidShapes(t *testing.T) {
	schema := write(t, "s.json", `{"type":"number"}`)
	for _, shape := range []string{"flag", "basic"} {
		code, out, errOut := do([]string{"validate", "-o", shape, schema}, `1`)
		if code != exitValid || out != `{"valid":true}`+"\n" || errOut != "" {
			t.Errorf("%s: code=%d out=%q err=%q", shape, code, out, errOut)
		}
	}
}

func TestStdinForms(t *testing.T) {
	schema := write(t, "s.json", `{"type":"number"}`)
	for _, args := range [][]string{{"validate", schema}, {"validate", schema, "-"}} {
		code, out, _ := do(args, `"x"`)
		if code != exitInvalid || !strings.HasPrefix(out, `{"valid":false`) {
			t.Errorf("%v: code=%d out=%q", args, code, out)
		}
	}
}

func TestUnknownOutputValue(t *testing.T) {
	schema := write(t, "s.json", `true`)
	doc := write(t, "d.json", `1`)
	for _, args := range [][]string{
		{"validate", "-o", "fancy", schema, doc},
		{"validate", "--output", "fancy", schema, doc},
		{"validate", "-o", "fancy", schema, doc, doc},
	} {
		code, out, errOut := do(args, "")
		wantErr(t, strings.Join(args, " "), code, out, errOut)
	}
}

func TestNoInstanceValueOnStderrOrStdout(t *testing.T) {
	schema := write(t, "s.json", `{"type":"number"}`)
	for _, in := range []string{`"` + secretValue + `"`, `{"` + secretValue + `":1,"` + secretValue + `":2}`} {
		_, out, errOut := do([]string{"validate", schema}, in)
		if strings.Contains(out, secretValue) || strings.Contains(errOut, secretValue) {
			t.Errorf("leak: out=%q err=%q", out, errOut)
		}
	}
}
