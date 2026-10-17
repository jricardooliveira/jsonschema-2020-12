package main

import (
	"fmt"
	"strings"
	"testing"
)

// Each limit flag, lowered so a small input exceeds it: exit 2, empty
// stdout, exactly one stderr line naming the flag.
func TestLimitFlagsContract(t *testing.T) {
	ref := writeTemp(t, `{"$id":"https://example.invalid/r.json","type":"object"}`)
	cases := []struct {
		flag, value, schema, doc string
		extra                    []string
	}{
		{"max-depth", "2", `true`, `[[[1]]]`, nil},
		{"max-input-bytes", "3", `true`, `   true`, nil},
		{"max-ref-chain", "5", `{"$ref":"#","type":"object"}`, `{}`, nil},
		{"max-steps", "1", `{"items":{"type":"integer"}}`, `[1,2,3,4,5]`, nil},
		{"max-pattern-size", "2", `{"pattern":"aaaa"}`, `"a"`, nil},
		{"max-output-units", "1", `{"items":{"minimum":5}}`, `[1,2,3]`, nil},
		{"max-report-bytes", "5", `{"minimum":5}`, `1`, nil},
		{"max-number-digits", "2", `true`, `12345`, nil},
		{"max-number-exponent", "2", `true`, `1e999`, nil},
		{"max-annotations", "1", `{"title":"t","description":"d"}`, `1`, nil},
		{"max-ref-file-bytes", "3", `true`, `1`, []string{"--ref", ref}},
		{"max-ref-total-bytes", "3", `true`, `1`, []string{"--ref", ref}},
	}
	if len(cases) != len(limitFlagNames) {
		t.Fatalf("%d cases for %d limit flags", len(cases), len(limitFlagNames))
	}
	for _, c := range cases {
		args := append([]string{"validate", "--" + c.flag, c.value}, c.extra...)
		args = append(args, writeTemp(t, c.schema))
		code, out, e := runCLI(t, c.doc, args...)
		want := "(raise with --" + c.flag + ")"
		if code != 2 || out != "" || !strings.HasPrefix(e, "limit exceeded: ") || !strings.HasSuffix(strings.TrimSpace(e), want) || strings.Count(strings.TrimSpace(e), "\n") != 0 {
			t.Errorf("%s: code %d out %q err %q", c.flag, code, out, e)
		}
	}
}

func TestLimitFlagZeroOrNegativeIsUsageError(t *testing.T) {
	sch := writeTemp(t, `true`)
	for _, name := range limitFlagNames {
		for _, v := range []string{"0", "-1"} {
			code, out, e := runCLI(t, "1", "validate", fmt.Sprintf("--%s=%s", name, v), sch)
			if code != 2 || out != "" || strings.TrimSpace(e) != "flag -"+name+" must be greater than 0" {
				t.Errorf("%s=%s: %d %q %q", name, v, code, out, e)
			}
		}
	}
}

// Several documents: the failing one gets an error line, the others still
// report, and exit 2 wins over invalid.
func TestLimitErrorInMultiDocumentRun(t *testing.T) {
	sch := writeTemp(t, `{"type":"string"}`)
	ok, bad, deep := writeTemp(t, `"x"`), writeTemp(t, `1`), writeTemp(t, `[[[1]]]`)
	code, out, e := runCLI(t, "", "validate", "--max-depth=2", sch, ok, deep, bad)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if code != 2 || e != "" || len(lines) != 3 {
		t.Fatalf("code %d lines %q err %q", code, lines, e)
	}
	if !strings.Contains(lines[1], `"error":"limit exceeded: nesting depth 2 (raise with --max-depth)"`) || strings.Contains(lines[1], `"output"`) {
		t.Errorf("error line: %s", lines[1])
	}
	if !strings.Contains(lines[0], `"output"`) || !strings.Contains(lines[2], `"output"`) {
		t.Errorf("report lines: %s / %s", lines[0], lines[2])
	}
}
