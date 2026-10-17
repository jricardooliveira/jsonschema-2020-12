package jsonschema

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Story 202: caching the built-in meta-schema must not move any hostile-input
// limit. Each case is the smallest input from cmd/jsonschema's
// TestLimitsAtTheEdge (measured before this story): at the edge the run gives
// its normal answer, one below it fails with the same limit error. The library
// is checked on a repeated Compile (the shared set is built by then) and the
// built CLI binary, a fresh process whose Compile is the first, must exit 2.

type cacheLimitCase struct {
	name, schema, doc, shape string
	limit                    Limit
	edge                     int
	set                      func(*Options, int)
	flag                     string
	wantCode                 int // CLI exit code at the edge
}

func cacheLimitCases() []cacheLimitCase {
	arr := func(n int, v string) string {
		return "[" + strings.TrimSuffix(strings.Repeat(v+",", n), ",") + "]"
	}
	var defs []string
	for i := 0; i < 4; i++ {
		defs = append(defs, fmt.Sprintf(`"r%d":{"$ref":"#/$defs/r%d"}`, i, i+1))
	}
	defs = append(defs, `"r4":true`)
	refSchema := `{"$defs":{` + strings.Join(defs, ",") + `},"$ref":"#/$defs/r0"}`
	return []cacheLimitCase{
		{"depth", `true`, `[[[[1]]]]`, "", LimitDepth, 4, func(o *Options, n int) { o.MaxDepth = n }, "max-depth", 0},
		{"ref chain", refSchema, `1`, "", LimitRefChain, 5, func(o *Options, n int) { o.MaxRefChain = n }, "max-ref-chain", 0},
		{"steps", `{"items":{"type":"integer"}}`, arr(5, "1"), "", LimitSteps, 186, func(o *Options, n int) { o.MaxSteps = n }, "max-steps", 0},
		{"output units", `{"items":{"type":"string"}}`, arr(4, "1"), "flag", LimitOutputUnits, 4, func(o *Options, n int) { o.MaxOutputUnits = n }, "max-output-units", 1},
		{"report bytes", `{"items":{"type":"string"}}`, arr(4, "1"), "basic", LimitReportBytes, 414, func(o *Options, n int) { o.MaxReportBytes = n }, "max-report-bytes", 1},
		{"annotations", `{"allOf":[{"title":"a"},{"title":"b"},{"title":"c"}]}`, `1`, "basic", LimitAnnotations, 37, func(o *Options, n int) { o.MaxAnnotations = n }, "max-annotations", 0},
	}
}

// runLib compiles with the limit set to n, validates and renders the shape.
func (c cacheLimitCase) runLib(n int) error {
	var o Options
	c.set(&o, n)
	s, err := CompileWithOptions([]byte(c.schema), o)
	if err != nil {
		return err
	}
	res, err := s.Validate([]byte(c.doc))
	if err != nil {
		return err
	}
	if c.shape != "" {
		_, err = res.Output(c.shape)
	}
	return err
}

func TestCacheLimitsLibrary(t *testing.T) {
	for _, c := range cacheLimitCases() {
		t.Run(c.name, func(t *testing.T) {
			for round := 0; round < 3; round++ { // later rounds reuse the cached meta-schema
				if err := c.runLib(c.edge); err != nil {
					t.Fatalf("round %d at the edge: %v", round, err)
				}
				err := c.runLib(c.edge - 1)
				var le *LimitError
				if !errors.As(err, &le) || le.Limit != c.limit {
					t.Fatalf("round %d one below: got %v, want limit %s", round, err, c.limit)
				}
			}
		})
	}
}

func TestCacheLimitsCLIFreshProcess(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "jsonschema")
	if out, err := exec.Command("go", "build", "-o", bin, "./cmd/jsonschema").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	for _, c := range cacheLimitCases() {
		t.Run(c.name, func(t *testing.T) {
			sch := filepath.Join(t.TempDir(), "schema.json")
			if err := os.WriteFile(sch, []byte(c.schema), 0o600); err != nil {
				t.Fatal(err)
			}
			try := func(n int) (int, string, string) {
				args := []string{"validate", "--" + c.flag + "=" + strconv.Itoa(n)}
				if c.shape != "" {
					args = append(args, "-o", c.shape)
				}
				cmd := exec.Command(bin, append(args, sch)...)
				cmd.Stdin = strings.NewReader(c.doc)
				var so, se bytes.Buffer
				cmd.Stdout, cmd.Stderr = &so, &se
				code := 0
				if err := cmd.Run(); err != nil {
					var ee *exec.ExitError
					if !errors.As(err, &ee) {
						t.Fatal(err)
					}
					code = ee.ExitCode()
				}
				return code, so.String(), se.String()
			}
			if code, _, e := try(c.edge); code != c.wantCode {
				t.Errorf("at the edge: code %d, want %d (stderr %q)", code, c.wantCode, e)
			}
			code, out, e := try(c.edge - 1)
			if code != 2 || out != "" || !strings.Contains(e, "limit exceeded") || !strings.Contains(e, "--"+c.flag) {
				t.Errorf("one below: code %d out %q stderr %q", code, out, e)
			}
		})
	}
}
