package applicator_test

import (
	"reflect"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/eval"
)

// The testdata/applicator/additionalproperties.json cases run through the root TestSpecCases.

func TestAdditionalPropertiesRecordsOnlyUncovered(t *testing.T) {
	n := mustCompile(t, `{"properties":{"a":true},"patternProperties":{"^p":true},"additionalProperties":true}`)
	_, rec, err := eval.EvaluateAnnotated(n, mustDecode(t, `{"a":1,"px":1,"z":1,"b":1}`))
	if err != nil {
		t.Fatal(err)
	}
	p := rec.PropertiesAt("")
	if !reflect.DeepEqual(p["/additionalProperties"], []string{"b", "z"}) ||
		!reflect.DeepEqual(p["/properties"], []string{"a"}) ||
		!reflect.DeepEqual(p["/patternProperties"], []string{"px"}) {
		t.Errorf("got %v", p)
	}
}

func TestAdditionalPropertiesMixedOrder(t *testing.T) {
	n := mustCompile(t, `{"properties":{"a":false},"dependentRequired":{"a":["m"]},"additionalProperties":false}`)
	_, units, err := eval.EvaluateUnits(n, mustDecode(t, `{"a":1,"z":1}`), 100)
	if err != nil {
		t.Fatal(err)
	}
	var kws []string
	for _, u := range units {
		kws = append(kws, u.KeywordLocation)
	}
	want := []string{"/additionalProperties", "/dependentRequired", "/properties"}
	var got []string
	for _, k := range kws {
		for _, w := range want {
			if len(k) >= len(w) && k[:len(w)] == w && (len(got) == 0 || got[len(got)-1] != w) {
				got = append(got, w)
			}
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("order %v from %v", got, kws)
	}
}
