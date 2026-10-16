package applicator_test

import (
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/eval"
	"github.com/neoface/jsonschema/internal/jsonvalue"
)

const secret = "s3cr3t-distinct-value"

func TestAllOfValidInvalid(t *testing.T) {
	n := mustCompile(t, `{"allOf":[{"type":"string"},true,{"const":"`+secret+`"}]}`)
	for doc, want := range map[string]bool{`"` + secret + `"`: true, `"x"`: false, `1`: false} {
		ok, err := eval.Evaluate(n, mustDecode(t, doc))
		if err != nil || ok != want {
			t.Errorf("%s: ok=%v err=%v want %v", doc, ok, err, want)
		}
	}
}

func TestAllOfBooleanSubschemas(t *testing.T) {
	for schema, want := range map[string]bool{`{"allOf":[true,true]}`: true, `{"allOf":[true,false]}`: false} {
		ok, err := eval.Evaluate(mustCompile(t, schema), mustDecode(t, `1`))
		if err != nil || ok != want {
			t.Errorf("%s: ok=%v err=%v", schema, ok, err)
		}
	}
}

func TestAllOfReportsEveryFailingSubschemaWithoutValues(t *testing.T) {
	n := mustCompile(t, `{"allOf":[{"type":"integer"},true,{"const":1}]}`)
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
	want := "/allOf/0/type,/allOf/2/const,/allOf"
	if strings.Join(locs, ",") != want {
		t.Errorf("got %v", locs)
	}
}

func TestAllOfSchemaErrors(t *testing.T) {
	for _, s := range []string{`{"allOf":[]}`, `{"allOf":{}}`, `{"allOf":true}`, `{"allOf":[1]}`, `{"allOf":["a"]}`} {
		v := mustDecode(t, s)
		_, err := compile.Compile(v)
		if err == nil || !strings.Contains(err.Error(), "allOf") {
			t.Errorf("%s: err=%v", s, err)
		}
	}
}

func TestAllOfChildErrorPropagates(t *testing.T) {
	n := mustCompile(t, `{"allOf":[{"type":"integer"},{"type":"integer"}]}`)
	ok, _, err := eval.EvaluateUnits(n, mustDecode(t, `"x"`), 1)
	var le *jsonvalue.LimitError
	if ok || err == nil || !asLimit(err, &le) {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}
