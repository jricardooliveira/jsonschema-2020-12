package main

import (
	"strings"
	"testing"
)

// Validation spec 6.3.3: a pattern the translator or RE2 cannot handle is a
// schema error (exit 2), never "invalid".
func TestUnsupportedPatternExitsTwo(t *testing.T) {
	for name, tc := range map[string]struct{ pattern, reason string }{
		"lookahead":                         {`(?=a)b`, "lookahead"},
		"lookbehind":                        {`(?<=a)b`, "lookbehind"},
		"backreference":                     {`(a)\\1`, "backreference"},
		"named backref":                     {`(?<n>a)\\k<n>`, "backreference"},
		"newline in pattern stays one line": {`\n(?=a)`, "lookahead"},
	} {
		s := write(t, "s.json", `{"pattern":"`+tc.pattern+`"}`)
		d := write(t, "d.json", `"SECRET-VALUE"`)
		code, out, errOut := do([]string{"validate", s, d}, "")
		if code != exitError || out != "" {
			t.Errorf("%s: code=%d out=%q", name, code, out)
		}
		if !strings.Contains(errOut, "pattern") || !strings.Contains(errOut, tc.reason) {
			t.Errorf("%s: stderr does not name keyword and problem: %q", name, errOut)
		}
		if strings.Count(strings.TrimRight(errOut, "\n"), "\n") != 0 || strings.Contains(errOut, "SECRET") {
			t.Errorf("%s: stderr not one value-free line: %q", name, errOut)
		}
	}
}
