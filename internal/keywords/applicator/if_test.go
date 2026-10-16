package applicator_test

import (
	"strings"
	"sync/atomic"
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/eval"
	"github.com/neoface/jsonschema/internal/keywords"
)

func TestIfThenElseValidity(t *testing.T) {
	const s = `{"if":{"type":"integer"},"then":{"minimum":10},"else":{"type":"string"}}`
	cases := map[string]bool{`20`: true, `5`: false, `"a"`: true, `1.5`: false}
	for doc, want := range cases {
		ok, err := eval.Evaluate(mustCompile(t, s), mustDecode(t, doc))
		if err != nil || ok != want {
			t.Errorf("%s: ok=%v err=%v", doc, ok, err)
		}
	}
}

func TestThenElseIgnoredWithoutIf(t *testing.T) {
	ok, err := eval.Evaluate(mustCompile(t, `{"then":false,"else":false}`), mustDecode(t, `1`))
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}

func TestIfAloneNeverFails(t *testing.T) {
	for _, s := range []string{`{"if":false}`, `{"if":{"type":"string"}}`, `{"if":false,"then":false}`} {
		ok, err := eval.Evaluate(mustCompile(t, s), mustDecode(t, `1`))
		if err != nil || !ok {
			t.Errorf("%s: ok=%v err=%v", s, ok, err)
		}
	}
}

func TestIfBooleanSchemas(t *testing.T) {
	for s, want := range map[string]bool{
		`{"if":true,"then":false}`: false, `{"if":true,"else":false}`: true,
		`{"if":false,"else":false}`: false, `{"if":false,"then":false}`: true,
	} {
		ok, err := eval.Evaluate(mustCompile(t, s), mustDecode(t, `1`))
		if err != nil || ok != want {
			t.Errorf("%s: ok=%v err=%v", s, ok, err)
		}
	}
}

func TestFailingBranchReportsBranchAndItsFailuresNotIf(t *testing.T) {
	for _, c := range []struct{ schema, doc, want string }{
		{`{"if":{"type":"string"},"then":{"minLength":99}}`, `"` + secret + `"`, "/then/minLength,/then"},
		{`{"if":{"type":"string"},"else":{"type":"integer"}}`, `"x"`, ""},
		{`{"if":{"type":"integer"},"else":{"const":1}}`, `"` + secret + `"`, "/else/const,/else"},
	} {
		ok, units, err := eval.EvaluateUnits(mustCompile(t, c.schema), mustDecode(t, c.doc), eval.DefaultMaxOutputUnits)
		if err != nil || ok != (c.want == "") {
			t.Fatalf("%s: ok=%v err=%v", c.schema, ok, err)
		}
		var locs []string
		for _, u := range units {
			locs = append(locs, u.KeywordLocation)
			if strings.Contains(u.Error, secret) {
				t.Errorf("leaks value: %q", u.Error)
			}
		}
		if got := strings.Join(locs, ","); got != c.want {
			t.Errorf("%s: got %q want %q", c.schema, got, c.want)
		}
	}
}

func TestIfThenElseSchemaErrors(t *testing.T) {
	for _, k := range []string{"if", "then", "else"} {
		for _, v := range []string{`1`, `"a"`, `[]`} {
			_, err := compile.Compile(mustDecode(t, `{"`+k+`":`+v+`}`))
			if err == nil || !strings.Contains(err.Error(), k) {
				t.Errorf("%s %s: err=%v", k, v, err)
			}
		}
	}
}

// x-count counts how often any schema holding it is compiled and evaluated.
var countCompiles, countEvals atomic.Int64

func init() {
	keywords.Register(keywords.Keyword{Name: "x-count", Vocabulary: "test",
		Compile:  func(*keywords.CompileContext, any) (any, error) { countCompiles.Add(1); return nil, nil },
		Evaluate: func(*keywords.EvalContext, any, any) (bool, error) { countEvals.Add(1); return true, nil }})
}

// TestNestedIfWorkIsLinear: 'if' is compiled and evaluated once per level, so
// work grows linearly with nesting (not 2^n or 3^n).
func TestNestedIfWorkIsLinear(t *testing.T) {
	const depth = 12
	leaf := `{"x-count":true}`
	inIf, inThen := leaf, leaf
	for i := 0; i < depth; i++ {
		inIf = `{"x-count":true,"if":` + inIf + `,"then":` + leaf + `,"else":` + leaf + `}`
		inThen = `{"x-count":true,"if":true,"then":` + inThen + `}`
	}
	for _, c := range []struct {
		name, schema          string
		compiles, evaluations int64
	}{
		// per level: itself, then, else compiled; itself and one branch evaluated
		{"nested in if", inIf, 3*depth + 1, 2*depth + 1},
		{"nested in then", inThen, depth + 1, depth + 1},
	} {
		countCompiles.Store(0)
		countEvals.Store(0)
		n := mustCompile(t, c.schema)
		if _, err := eval.Evaluate(n, mustDecode(t, `1`)); err != nil {
			t.Fatal(err)
		}
		if countCompiles.Load() != c.compiles || countEvals.Load() != c.evaluations {
			t.Errorf("%s: compiles=%d evals=%d, want %d %d", c.name,
				countCompiles.Load(), countEvals.Load(), c.compiles, c.evaluations)
		}
	}
}

func TestIfThenElseChildErrorPropagates(t *testing.T) {
	// the discarded 'if' unit alone exceeds a cap of 0: error, not valid/invalid
	n := mustCompile(t, `{"if":{"type":"string"},"then":true}`)
	ok, _, err := eval.EvaluateUnits(n, mustDecode(t, `1`), 0)
	if ok || err == nil {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}
