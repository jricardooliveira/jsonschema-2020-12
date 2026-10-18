package core_test

import (
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/eval"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords/core"
)

func compiled(t *testing.T, src string) *compile.Node {
	t.Helper()
	v, err := jsonvalue.DecodeLimited([]byte(src), jsonvalue.NumberLimits{})
	if err != nil {
		t.Fatal(err)
	}
	n, err := compile.Compile(v)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// refNode compiles a schema and points its $ref at target (what load-time
// resolution does).
func refNode(t *testing.T, schema, target string) *compile.Node {
	n := compiled(t, schema)
	for _, it := range n.Items {
		if r, ok := it.Data.(*core.Ref); ok {
			r.Target = compiled(t, target)
		}
	}
	return n
}

// core 8.2.3.1: $ref applies alongside siblings; both must pass.
func TestRefWithSiblings(t *testing.T) {
	n := refNode(t, `{"$ref":"#/x","type":"string"}`, `{"minLength":2}`)
	for doc, want := range map[string]bool{`"ab"`: true, `"a"`: false, `5`: false} {
		d, _ := jsonvalue.DecodeLimited([]byte(doc), jsonvalue.NumberLimits{})
		ok, err := eval.Evaluate(n, d)
		if err != nil || ok != want {
			t.Errorf("%s: %v %v", doc, ok, err)
		}
	}
}

func TestUnresolvedRefIsError(t *testing.T) {
	n := compiled(t, `{"$ref":"#/nope"}`)
	d, _ := jsonvalue.DecodeLimited([]byte(`1`), jsonvalue.NumberLimits{})
	if ok, err := eval.Evaluate(n, d); ok || err == nil {
		t.Errorf("%v %v", ok, err)
	}
}
