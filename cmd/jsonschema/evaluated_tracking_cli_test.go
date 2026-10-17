package main

import (
	"sort"
	"strconv"
	"strings"
	"testing"
)

// withNote adds an unknown keyword holding the name kw. The load-time scan
// counts "unevaluatedItems" anywhere, so that variant keeps recording; the
// "unevaluatedNot" variant skips it (see evaluated_tracking_limits_test.go).
func withNote(schema, kw string) string {
	note := `"x-note":{"` + kw + `":true}`
	if schema == "true" {
		return "{" + note + "}"
	}
	return strings.TrimSuffix(schema, "}") + "," + note + "}"
}

// TestEvaluatedTrackingLimitsCLI: through the CLI, every limit's boundary is
// the same on a schema that skips the evaluated record and one that keeps it:
// below the boundary the run exits 2 with the same limit error, at it the
// run gives a normal answer.
func TestEvaluatedTrackingLimitsCLI(t *testing.T) {
	for _, c := range edgeCases() {
		t.Run(c.name, func(t *testing.T) {
			type edge struct {
				n      int
				stderr string
				atCode int
			}
			find := func(schema string) edge {
				sch := writeTemp(t, schema)
				try := func(n int) (int, string, string) {
					args := []string{"validate", "--" + c.flag + "=" + strconv.Itoa(n)}
					if c.shape != "" {
						args = append(args, "-o", c.shape)
					}
					return runCLI(t, c.doc, append(args, sch)...)
				}
				const hi = 100000
				n := 1 + sort.Search(hi, func(i int) bool { code, _, _ := try(1 + i); return code != 2 })
				if n > hi || n < 2 {
					t.Fatalf("no boundary for %s in 2..%d (got %d)", schema, hi, n)
				}
				atCode, _, _ := try(n)
				code, out, e := try(n - 1)
				if code != 2 || out != "" || !strings.Contains(e, "limit exceeded") || !strings.Contains(e, "--"+c.flag) {
					t.Fatalf("one over: code %d out %q stderr %q", code, out, e)
				}
				return edge{n, e, atCode}
			}
			skip := find(withNote(c.schema, "unevaluatedNot"))
			keep := find(withNote(c.schema, "unevaluatedItems"))
			if skip != keep {
				t.Errorf("skipping %+v, recording %+v", skip, keep)
			}
			if skip.atCode != c.wantCode {
				t.Errorf("at the limit: code %d, want %d", skip.atCode, c.wantCode)
			}
		})
	}
}
