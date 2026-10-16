package applicator_test

import (
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/eval"
)

func TestAnyOfValidInvalid(t *testing.T) {
	cases := []struct {
		schema, doc string
		want        bool
	}{
		{`{"anyOf":[{"type":"string"},{"type":"integer"}]}`, `1`, true},
		{`{"anyOf":[{"type":"string"},{"type":"integer"}]}`, `1.5`, false},
		{`{"anyOf":[false,true]}`, `1`, true},
		{`{"anyOf":[false,false]}`, `1`, false},
	}
	for _, c := range cases {
		ok, err := eval.Evaluate(mustCompile(t, c.schema), mustDecode(t, c.doc))
		if err != nil || ok != c.want {
			t.Errorf("%s %s: ok=%v err=%v", c.schema, c.doc, ok, err)
		}
	}
}

func TestAnyOfEvaluatesEveryChoice(t *testing.T) {
	// choice 0 matches; choice 1 still runs: its unit trips a cap of 0.
	n := mustCompile(t, `{"anyOf":[true,{"type":"integer"}]}`)
	ok, _, err := eval.EvaluateUnits(n, mustDecode(t, `"x"`), 0)
	if ok || err == nil {
		t.Fatalf("later choice not evaluated: ok=%v err=%v", ok, err)
	}
}

func TestAnyOfReportsEachChoiceWithoutValues(t *testing.T) {
	n := mustCompile(t, `{"anyOf":[{"type":"integer"},{"const":1}]}`)
	ok, units, err := eval.EvaluateUnits(n, mustDecode(t, `"`+secret+`"`), eval.DefaultMaxOutputUnits)
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	var locs []string
	for _, u := range units {
		locs = append(locs, u.KeywordLocation)
		if strings.Contains(u.Error, secret) {
			t.Errorf("leaks value: %q", u.Error)
		}
	}
	if got := strings.Join(locs, ","); got != "/anyOf/0/type,/anyOf/1/const,/anyOf" {
		t.Errorf("got %s", got)
	}
}

func TestAnyOfSchemaErrors(t *testing.T) {
	for _, s := range []string{`{"anyOf":[]}`, `{"anyOf":{}}`, `{"anyOf":true}`, `{"anyOf":[1]}`} {
		_, err := compile.Compile(mustDecode(t, s))
		if err == nil || !strings.Contains(err.Error(), "anyOf") {
			t.Errorf("%s: err=%v", s, err)
		}
	}
}
