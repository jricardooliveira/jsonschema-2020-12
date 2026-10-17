package core_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/eval"
	"github.com/neoface/jsonschema/internal/jsonvalue"
	_ "github.com/neoface/jsonschema/internal/keywords/all"
)

const treeSchema = `{"$id":"https://t.example/tree","$dynamicAnchor":"node","type":"object",
 "properties":{"data":true,"children":{"type":"array","items":{"$dynamicRef":"#node"}}}}`

// core 8.2.3.2 recursive extension: the extension's `required` applies at
// every level; the plain tree does not require it.
const strictSchema = `{"$id":"https://t.example/strict","$dynamicAnchor":"node","$ref":"tree",
 "required":["data"],"$defs":{"tree":` + treeSchema + `}}`

func runDyn(t *testing.T, schema, doc string) (bool, error) {
	t.Helper()
	s, err := jsonvalue.DecodeLimited([]byte(schema), jsonvalue.NumberLimits{})
	if err != nil {
		t.Fatal(err)
	}
	d, err := jsonvalue.DecodeLimited([]byte(doc), jsonvalue.NumberLimits{})
	if err != nil {
		t.Fatal(err)
	}
	n, err := compile.CompileDocument(s, "file:///s.json")
	if err != nil {
		t.Fatal(err)
	}
	return eval.Evaluate(n, d)
}

func TestDynamicRefRecursiveExtension(t *testing.T) {
	const deepBad = `{"data":1,"children":[{"data":2,"children":[{"children":[]}]}]}`
	const deepOK = `{"data":1,"children":[{"data":2,"children":[{"data":3}]}]}`
	for name, c := range map[string]struct {
		schema, doc string
		want        bool
	}{
		"plain tree, deep missing data": {treeSchema, deepBad, true},
		"extension, deep ok":            {strictSchema, deepOK, true},
		"extension, deep missing data":  {strictSchema, deepBad, false},
	} {
		if ok, err := runDyn(t, c.schema, c.doc); err != nil || ok != c.want {
			t.Errorf("%s: %v %v", name, ok, err)
		}
	}
}

// core 8.2.3.2 bookending: no $dynamicAnchor at the target means $ref.
func TestDynamicRefBookending(t *testing.T) {
	s := `{"$id":"https://t.example/b","$dynamicAnchor":"outer","$defs":{"x":{"$anchor":"x","type":"string"}},"$dynamicRef":"#x"}`
	for doc, want := range map[string]bool{`"a"`: true, `1`: false} {
		if ok, err := runDyn(t, s, doc); err != nil || ok != want {
			t.Errorf("%s: %v %v", doc, ok, err)
		}
	}
}

func TestDynamicRefUnresolvedIsSchemaError(t *testing.T) {
	s, _ := jsonvalue.DecodeLimited([]byte(`{"$dynamicRef":"#nope"}`), jsonvalue.NumberLimits{})
	if _, err := compile.CompileDocument(s, "file:///s.json"); err == nil {
		t.Fatal("want unresolved-reference error")
	}
}

func TestDynamicRefLocationNamesKeyword(t *testing.T) {
	s, _ := jsonvalue.DecodeLimited([]byte(`{"$defs":{"a":{"$anchor":"a","type":"string"}},"$dynamicRef":"#a"}`), jsonvalue.NumberLimits{})
	d, _ := jsonvalue.DecodeLimited([]byte(`1`), jsonvalue.NumberLimits{})
	n, err := compile.CompileDocument(s, "file:///s.json")
	if err != nil {
		t.Fatal(err)
	}
	ok, units, err := eval.EvaluateUnits(n, d, eval.DefaultMaxOutputUnits)
	if err != nil || ok || len(units) == 0 {
		t.Fatalf("%v %v %v", ok, units, err)
	}
	found := false
	for _, u := range units {
		found = found || strings.HasPrefix(u.KeywordLocation, "/$dynamicRef")
	}
	if !found {
		t.Errorf("no unit under /$dynamicRef: %+v", units)
	}
}

// A dynamic cycle that never descends is a loop error, not valid or invalid.
func TestDynamicRefCycleIsLoopError(t *testing.T) {
	s := `{"$defs":{"a":{"$dynamicAnchor":"a","$dynamicRef":"#a"}},"$ref":"#/$defs/a"}`
	ok, err := runDyn(t, s, `{}`)
	var le *eval.LoopError
	if ok || !errors.As(err, &le) {
		t.Errorf("%v %v", ok, err)
	}
}
