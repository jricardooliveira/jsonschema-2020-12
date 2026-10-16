package applicator_test

import (
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/eval"
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

func TestIfThenElseChildErrorPropagates(t *testing.T) {
	// the discarded 'if' unit alone exceeds a cap of 0: error, not valid/invalid
	n := mustCompile(t, `{"if":{"type":"string"},"then":true}`)
	ok, _, err := eval.EvaluateUnits(n, mustDecode(t, `1`), 0)
	if ok || err == nil {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}
