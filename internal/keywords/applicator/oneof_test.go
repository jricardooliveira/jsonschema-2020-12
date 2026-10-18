package applicator_test

import (
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/eval"
)

func TestOneOfValidInvalid(t *testing.T) {
	cases := []struct {
		schema, doc string
		want        bool
	}{
		{`{"oneOf":[{"type":"string"},{"type":"integer"}]}`, `1`, true},
		{`{"oneOf":[{"type":"string"},{"type":"integer"}]}`, `1.5`, false},
		{`{"oneOf":[true,true]}`, `1`, false},
		{`{"oneOf":[false,true]}`, `1`, true},
	}
	for _, c := range cases {
		ok, err := eval.Evaluate(mustCompile(t, c.schema), mustDecode(t, c.doc))
		if err != nil || ok != c.want {
			t.Errorf("%s %s: ok=%v err=%v", c.schema, c.doc, ok, err)
		}
	}
}

func TestOneOfEvaluatesEveryChoice(t *testing.T) {
	n := mustCompile(t, `{"oneOf":[true,true,{"type":"integer"}]}`)
	ok, _, err := eval.EvaluateUnits(n, mustDecode(t, `"x"`), 0)
	if ok || err == nil {
		t.Fatalf("later choice not evaluated: ok=%v err=%v", ok, err)
	}
}

func TestOneOfNamesMatchingChoices(t *testing.T) {
	n := mustCompile(t, `{"oneOf":[true,{"type":"integer"},{"const":"`+secret+`"}]}`)
	_, units, err := eval.EvaluateUnits(n, mustDecode(t, `"`+secret+`"`), eval.DefaultMaxOutputUnits)
	if err != nil {
		t.Fatal(err)
	}
	// choice 1 fails (one unit), then oneOf's own unit.
	last := units[len(units)-1]
	if last.KeywordLocation != "/oneOf" || last.Error != "oneOf: choices 0 and 2 both matched" {
		t.Errorf("unit %+v", last)
	}
	n = mustCompile(t, `{"oneOf":[true,true,true]}`)
	_, units, _ = eval.EvaluateUnits(n, mustDecode(t, `1`), eval.DefaultMaxOutputUnits)
	if len(units) != 1 || units[0].Error != "oneOf: choices 0, 1 and 2 all matched" {
		t.Errorf("units %+v", units)
	}
}

func TestOneOfNoMatchReportsChoicesWithoutValues(t *testing.T) {
	n := mustCompile(t, `{"oneOf":[{"type":"integer"},{"const":1}]}`)
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
	if got := strings.Join(locs, ","); got != "/oneOf/0/type,/oneOf/1/const,/oneOf" {
		t.Errorf("got %s", got)
	}
}

func TestOneOfSchemaErrors(t *testing.T) {
	for _, s := range []string{`{"oneOf":[]}`, `{"oneOf":{}}`, `{"oneOf":[1]}`} {
		_, err := compile.Compile(mustDecode(t, s))
		if err == nil || !strings.Contains(err.Error(), "oneOf") {
			t.Errorf("%s: err=%v", s, err)
		}
	}
}
