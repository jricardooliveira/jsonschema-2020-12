package jsonschema

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"
)

// trackSchemas returns two schemas with the same shape and cost: skip has no
// unevaluated* keyword, so recording is off; keep names one inside an unknown
// keyword, which the cautious load-time scan counts, so recording stays on.
func trackSchemas(body string) (skip, keep string) {
	tpl := `{%s,"x-note":{"%s":true}}`
	return fmt.Sprintf(tpl, body, "unevaluatedNot"), fmt.Sprintf(tpl, body, "unevaluatedItems")
}

// limitProbe compiles with the limit set to n, then validates (and renders).
type limitProbe func(n int) (compileErr, runErr error)

// boundary binary-searches the smallest n in (lo, hi] for which the run
// succeeds, then checks that n-1 compiles and fails with the wanted limit.
// It returns n and the error text at n-1.
func boundary(t *testing.T, want Limit, lo, hi int, probe limitProbe) (int, string) {
	t.Helper()
	ok := func(n int) bool { c, r := probe(n); return c == nil && r == nil }
	if ok(lo) || !ok(hi) {
		t.Fatalf("limit %s: bounds %d..%d do not bracket the boundary", want, lo, hi)
	}
	n := lo + 1 + sort.Search(hi-lo, func(i int) bool { return ok(lo + 1 + i) })
	c, r := probe(n - 1)
	var le *LimitError
	if c != nil || !errors.As(r, &le) || le.Limit != want {
		t.Fatalf("limit %s at %d: compile %v, run %v", want, n-1, c, r)
	}
	return n, r.Error()
}

func TestEvaluatedTrackingLimitsUnchanged(t *testing.T) {
	deep := strings.Repeat("[", 20) + strings.Repeat("]", 20)
	many := func(item string, n int) string {
		return "[" + strings.TrimSuffix(strings.Repeat(item+",", n), ",") + "]"
	}
	var chain strings.Builder
	chain.WriteString(`"$ref":"#/$defs/r0","$defs":{`)
	for i := 0; i < 30; i++ {
		fmt.Fprintf(&chain, `"r%d":{"$ref":"#/$defs/r%d"},`, i, i+1)
	}
	chain.WriteString(`"r30":true}`)

	set := func(f func(*Options, int)) func(string, string) limitProbe {
		return func(schema, doc string) limitProbe {
			return func(n int) (error, error) {
				var o Options
				f(&o, n)
				s, err := CompileWithOptions([]byte(schema), o)
				if err != nil {
					return err, nil
				}
				_, err = s.Validate([]byte(doc))
				return nil, err
			}
		}
	}
	render := func(shape string) func(string, string) limitProbe {
		return func(schema, doc string) limitProbe {
			return func(n int) (error, error) {
				s, err := CompileWithOptions([]byte(schema), Options{MaxReportBytes: n})
				if err != nil {
					return err, nil
				}
				res, err := s.Validate([]byte(doc))
				if err != nil {
					return err, nil
				}
				_, err = res.Output(shape)
				return nil, err
			}
		}
	}
	failing := many(`1`, 20)
	cases := []struct {
		name, body, doc string
		want            Limit
		lo, hi          int
		probe           func(string, string) limitProbe
	}{
		{"depth", `"type":"array"`, deep, LimitDepth, 4, 100, set(func(o *Options, n int) { o.MaxDepth = n })},
		{"ref chain", chain.String(), `1`, LimitRefChain, 20, 80, set(func(o *Options, n int) { o.MaxRefChain = n })},
		{"steps", `"items":{"type":"integer"}`, many(`1`, 500), LimitSteps, 50, 100000, set(func(o *Options, n int) { o.MaxSteps = n })},
		{"output units", `"items":{"type":"string"}`, failing, LimitOutputUnits, 1, 100, set(func(o *Options, n int) { o.MaxOutputUnits = n })},
		{"annotations", `"items":{"title":"t"}`, many(`1`, 50), LimitAnnotations, 1, 1000, set(func(o *Options, n int) { o.MaxAnnotations = n })},
		{"report bytes basic", `"items":{"type":"string"}`, failing, LimitReportBytes, 1, 100000, render("basic")},
		{"report bytes verbose", `"items":{"type":"string"}`, failing, LimitReportBytes, 1, 1000000, render("verbose")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			skip, keep := trackSchemas(c.body)
			for schema, wantSkip := range map[string]bool{skip: true, keep: false} {
				s, err := Compile([]byte(schema))
				if err != nil {
					t.Fatal(err)
				}
				if s.root.NoUnevaluated != wantSkip {
					t.Fatalf("NoUnevaluated = %v, want %v", s.root.NoUnevaluated, wantSkip)
				}
			}
			ns, es := boundary(t, c.want, c.lo, c.hi, c.probe(skip, c.doc))
			nk, ek := boundary(t, c.want, c.lo, c.hi, c.probe(keep, c.doc))
			if ns != nk || es != ek {
				t.Errorf("skipping boundary %d (%q), recording boundary %d (%q)", ns, es, nk, ek)
			}
		})
	}
}
