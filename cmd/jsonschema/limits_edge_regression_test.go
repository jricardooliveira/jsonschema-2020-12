package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// edgeCase fixes a small generated schema and document and one limit flag.
// edge is the smallest flag value that still gives a normal answer; it was
// measured on the code before locations became lazy. edge-1 must give exit 2
// with a limit error, edge must give wantCode.
type edgeCase struct {
	name, flag, schema, doc, shape string
	edge, wantCode                 int
}

func edgeCases() []edgeCase {
	arr := func(n int, v string) string {
		return "[" + strings.TrimSuffix(strings.Repeat(v+",", n), ",") + "]"
	}
	// $ref chain: #/$defs/r0 -> r1 -> ... -> r4 -> true
	var defs []string
	for i := 0; i < 4; i++ {
		defs = append(defs, fmt.Sprintf(`"r%d":{"$ref":"#/$defs/r%d"}`, i, i+1))
	}
	defs = append(defs, `"r4":true`)
	refSchema := `{"$defs":{` + strings.Join(defs, ",") + `},"$ref":"#/$defs/r0"}`
	titles := `{"allOf":[{"title":"a"},{"title":"b"},{"title":"c"}]}`
	return []edgeCase{
		{"depth", "max-depth", `true`, `[[[[1]]]]`, "", 4, 0},
		{"ref chain", "max-ref-chain", refSchema, `1`, "", 5, 0},
		{"steps", "max-steps", `{"items":{"type":"integer"}}`, arr(5, "1"), "", 186, 0},
		{"output units", "max-output-units", `{"items":{"type":"string"}}`, arr(4, "1"), "flag", 4, 1},
		{"report bytes", "max-report-bytes", `{"items":{"type":"string"}}`, arr(4, "1"), "basic", 414, 1},
		{"annotations", "max-annotations", titles, `1`, "basic", 37, 0},
	}
}

func TestLimitsAtTheEdge(t *testing.T) {
	for _, c := range edgeCases() {
		t.Run(c.name, func(t *testing.T) {
			sch := writeTemp(t, c.schema)
			try := func(n int) (int, string, string) {
				args := []string{"validate", "--" + c.flag + "=" + strconv.Itoa(n)}
				if c.shape != "" {
					args = append(args, "-o", c.shape)
				}
				return runCLI(t, c.doc, append(args, sch)...)
			}
			if code, _, e := try(c.edge); code != c.wantCode {
				t.Errorf("at the limit: code %d, want %d (stderr %q)", code, c.wantCode, e)
			}
			code, out, e := try(c.edge - 1)
			if code != 2 || out != "" || !strings.Contains(e, "limit exceeded") || !strings.Contains(e, "--"+c.flag) {
				t.Errorf("one over: code %d out %q stderr %q", code, out, e)
			}
		})
	}
}
