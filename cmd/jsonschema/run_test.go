package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
)

func write(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func do(args []string, stdin string) (int, string, string) {
	var o, e bytes.Buffer
	code := run(args, strings.NewReader(stdin), &o, &e)
	return code, o.String(), e.String()
}

func wantErr(t *testing.T, name string, code int, out, errOut string) {
	t.Helper()
	if code != exitError || out != "" || strings.Count(errOut, "\n") != 1 || !strings.HasSuffix(errOut, "\n") {
		t.Errorf("%s: code=%d out=%q err=%q", name, code, out, errOut)
	}
}

func TestExitValidInvalid(t *testing.T) {
	tr := write(t, "t.json", `true`)
	fl := write(t, "f.json", `false`)
	doc := write(t, "d.json", `{"a":1}`)
	cases := []struct {
		args  []string
		stdin string
		code  int
		valid bool
	}{
		{[]string{"validate", tr, doc}, "", exitValid, true},
		{[]string{"validate", fl, doc}, "", exitInvalid, false},
		{[]string{"validate", tr}, `[1]`, exitValid, true},
		{[]string{"validate", fl, "-"}, `[1]`, exitInvalid, false},
	}
	for _, c := range cases {
		code, out, errOut := do(c.args, c.stdin)
		var got map[string]any
		if err := json.Unmarshal([]byte(out), &got); err != nil || got["valid"] != any(c.valid) || strings.Count(out, "\n") != 1 || !strings.HasSuffix(out, "\n") {
			t.Errorf("%v: bad stdout %q", c.args, out)
		}
		if code != c.code || errOut != "" {
			t.Errorf("%v: code=%d err=%q", c.args, code, errOut)
		}
	}
}

func TestExitErrors(t *testing.T) {
	tr := write(t, "t.json", `true`)
	doc := write(t, "d.json", `{}`)
	five := write(t, "five.json", `5`)
	broken := write(t, "b.json", `{`)
	missing := filepath.Join(t.TempDir(), "nope.json")
	cases := map[string]struct {
		args  []string
		stdin string
	}{
		"no args":          {nil, ""},
		"unknown sub":      {[]string{"frobnicate"}, ""},
		"missing schema":   {[]string{"validate"}, ""},
		"unknown flag":     {[]string{"validate", "-bogus", tr}, ""},
		"help":             {[]string{"validate", "-h"}, ""},
		"schema missing":   {[]string{"validate", missing, doc}, ""},
		"doc missing":      {[]string{"validate", tr, missing}, ""},
		"schema broken":    {[]string{"validate", broken, doc}, ""},
		"doc broken":       {[]string{"validate", tr, broken}, ""},
		"stdin broken":     {[]string{"validate", tr}, `{`},
		"stdin dash empty": {[]string{"validate", tr, "-"}, ``},
		"schema not obj":   {[]string{"validate", five, doc}, ""},
		"deep doc":         {[]string{"validate", tr}, strings.Repeat("[", 100000)},
		"deep doc limit":   {[]string{"validate", tr}, strings.Repeat("[", jsonvalue.MaxDepth+1)},
		"oversize":         {[]string{"validate", tr}, strings.Repeat(" ", maxInputBytes+1)},
	}
	for name, c := range cases {
		code, out, errOut := do(c.args, c.stdin)
		wantErr(t, name, code, out, errOut)
	}
}

func TestStderrNoContent(t *testing.T) {
	tr := write(t, "t.json", `true`)
	for _, in := range []string{`{"pw": "s3cr3t" !}`, `{"s3cr3t":1,"s3cr3t":2}`, `s3cr3t`} {
		code, out, errOut := do([]string{"validate", tr}, in)
		wantErr(t, in, code, out, errOut)
		if strings.Contains(errOut, "s3cr3t") {
			t.Errorf("%s: leaked: %q", in, errOut)
		}
	}
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("x") }

func TestWriteFailure(t *testing.T) {
	tr := write(t, "t.json", `true`)
	var e bytes.Buffer
	if code := run([]string{"validate", tr}, strings.NewReader(`1`), failWriter{}, &e); code != exitError {
		t.Errorf("code=%d", code)
	}
}
