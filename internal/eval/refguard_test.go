package eval_test

import (
	"errors"
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/eval"
	"github.com/neoface/jsonschema/internal/jsonvalue"
	_ "github.com/neoface/jsonschema/internal/keywords/all"
)

func guardRun(t *testing.T, schema, doc string) (bool, error) {
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

// core 8.2.3.1: a reference that descends into the document is not a loop.
func TestTreeRecursionIsNotALoop(t *testing.T) {
	const tree = `{"type":"object","properties":{"v":{"type":"integer"},"kid":{"$ref":"#"}}}`
	for doc, want := range map[string]bool{
		`{"v":1,"kid":{"v":2,"kid":{"v":3}}}`: true,
		`{"v":1,"kid":{"v":"x"}}`:             false,
	} {
		if ok, err := guardRun(t, tree, doc); err != nil || ok != want {
			t.Errorf("%s: %v %v", doc, ok, err)
		}
	}
}

// Re-entering the same schema at the same instance location is a loop error.
func TestNonDescendingLoop(t *testing.T) {
	ok, err := guardRun(t, `{"$ref":"#","type":"object"}`, `{}`)
	var le *eval.LoopError
	if ok || !errors.As(err, &le) || le.SchemaLocation != "" || le.InstanceLocation != "" {
		t.Errorf("%v %v", ok, err)
	}
}

// The same subschema reached twice in sequence (not nested) is fine.
func TestRepeatedSiblingUseIsNotALoop(t *testing.T) {
	s := `{"$defs":{"s":{"type":"string"}},"properties":{"a":{"$ref":"#/$defs/s"},"b":{"$ref":"#/$defs/s"}},"allOf":[{"$ref":"#/$defs/s"}]}`
	if ok, err := guardRun(t, s, `"x"`); err != nil || !ok {
		t.Errorf("%v %v", ok, err)
	}
	if ok, err := guardRun(t, s, `{"a":"x","b":"y"}`); err != nil || ok {
		t.Errorf("%v %v", ok, err)
	}
}
