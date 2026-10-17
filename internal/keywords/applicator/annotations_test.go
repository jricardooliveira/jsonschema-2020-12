package applicator_test

import (
	"reflect"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/eval"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	_ "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords/all"
)

func annotate(t *testing.T, schema, doc string) map[string]any {
	t.Helper()
	lim := jsonvalue.NumberLimits{MaxDigits: 1000, MaxExponent: 1000}
	s, err := jsonvalue.DecodeLimited([]byte(schema), lim)
	if err != nil {
		t.Fatal(err)
	}
	n, err := compile.Compile(s)
	if err != nil {
		t.Fatal(err)
	}
	d, err := jsonvalue.DecodeLimited([]byte(doc), lim)
	if err != nil {
		t.Fatal(err)
	}
	ev, err := eval.EvaluateFull(n, d, eval.DefaultMaxOutputUnits, eval.DefaultMaxAnnotations)
	if err != nil || !ev.Valid {
		t.Fatal(ev.Valid, err)
	}
	out := map[string]any{}
	for _, a := range ev.Annotations {
		out[a.Keyword] = a.Value
	}
	return out
}

func TestApplicatorAnnotations(t *testing.T) {
	cases := []struct {
		name, schema, doc, kw string
		want                  any
	}{
		{"properties", `{"properties":{"a":true,"c":true}}`, `{"a":1,"b":2}`, "properties", []string{"a"}},
		{"patternProperties", `{"patternProperties":{"^b":true}}`, `{"a":1,"b":2}`, "patternProperties", []string{"b"}},
		{"additionalProperties", `{"properties":{"a":true},"additionalProperties":true}`, `{"a":1,"b":2}`, "additionalProperties", []string{"b"}},
		{"prefixItems largest", `{"prefixItems":[true,true]}`, `[1,2,3]`, "prefixItems", 1},
		{"prefixItems all", `{"prefixItems":[true,true]}`, `[1,2]`, "prefixItems", true},
		{"items", `{"prefixItems":[true],"items":true}`, `[1,2]`, "items", true},
		{"contains indices", `{"contains":{"type":"integer"}}`, `[1,"a",2]`, "contains", []int{0, 2}},
		{"contains all", `{"contains":{"type":"integer"}}`, `[1,2]`, "contains", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := annotate(t, c.schema, c.doc)[c.kw]; !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %#v want %#v", got, c.want)
			}
		})
	}
}
