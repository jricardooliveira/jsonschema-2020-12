package applicator_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/eval"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
)

func asLimit(err error, target **jsonvalue.LimitError) bool { return errors.As(err, target) }

func TestNotValidInvalidAndBooleans(t *testing.T) {
	cases := []struct {
		schema, doc string
		want        bool
	}{
		{`{"not":{"type":"string"}}`, `1`, true},
		{`{"not":{"type":"string"}}`, `"a"`, false},
		{`{"not":true}`, `1`, false},
		{`{"not":false}`, `1`, true},
	}
	for _, c := range cases {
		ok, err := eval.Evaluate(mustCompile(t, c.schema), mustDecode(t, c.doc))
		if err != nil || ok != c.want {
			t.Errorf("%s %s: ok=%v err=%v", c.schema, c.doc, ok, err)
		}
	}
}

func TestNotReportsOneUnitWithoutValues(t *testing.T) {
	n := mustCompile(t, `{"not":{"const":"`+secret+`"}}`)
	ok, units, err := eval.EvaluateUnits(n, mustDecode(t, `"`+secret+`"`), eval.DefaultMaxOutputUnits)
	if err != nil || ok || len(units) != 1 {
		t.Fatalf("ok=%v err=%v units=%+v", ok, err, units)
	}
	if units[0].KeywordLocation != "/not" || strings.Contains(units[0].Error, secret) {
		t.Errorf("unit %+v", units[0])
	}
}

func TestNotChildErrorIsNotAPass(t *testing.T) {
	// the passing child's discarded unit still counts against the cap
	n := mustCompile(t, `{"not":{"type":"integer"}}`)
	ok, _, err := eval.EvaluateUnits(n, mustDecode(t, `"x"`), 0)
	if ok || err == nil {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}

func TestNotSchemaError(t *testing.T) {
	for _, s := range []string{`{"not":1}`, `{"not":"a"}`, `{"not":[]}`} {
		_, err := compile.Compile(mustDecode(t, s))
		if err == nil || !strings.Contains(err.Error(), "not") {
			t.Errorf("%s: err=%v", s, err)
		}
	}
}
