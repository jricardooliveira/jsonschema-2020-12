package eval_test

import (
	"errors"
	"fmt"
	"strings"
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

func nested(open, mid, close string, n int) string {
	return strings.Repeat(open, n) + mid + strings.Repeat(close, n)
}

// The depth limit is the backstop for recursion that descends: a hostile
// schema gets a LimitError (exit 2), never a stack overflow.
func TestEvalDepthLimit(t *testing.T) {
	k := 10 // allOf wrappers per instance level: k+2 nodes per level
	schema := `{"items":` + nested(`{"allOf":[`, `{"$ref":"#"}`, `]}`, k) + `}`
	doc := nested("[", "", "]", eval.MaxEvalDepth/(k+2)+10)
	ok, err := guardRun(t, schema, doc)
	var le *jsonvalue.LimitError
	if ok || !errors.As(err, &le) || le.Max != jsonvalue.MaxDepth { // the nesting depth that was set
		t.Errorf("%v %v", ok, err)
	}
}

// A legitimate recursive schema over the deepest document the decoder
// accepts stays well inside the limit.
func TestEvalDepthLimitIsGenerous(t *testing.T) {
	doc := nested("[", "", "]", jsonvalue.MaxDepth)
	if ok, err := guardRun(t, `{"items":{"$ref":"#"}}`, doc); err != nil || !ok {
		t.Errorf("%v %v", ok, err)
	}
}

// keywordLocation follows the evaluation path through $ref (core 12.3.1).
func TestRefKeywordLocationIsEvaluationPath(t *testing.T) {
	s, _ := jsonvalue.DecodeLimited([]byte(`{"$defs":{"a":{"minLength":2}},"properties":{"p":{"$ref":"#/$defs/a"}},"then":{"$ref":"#/$defs/a"},"if":true}`), jsonvalue.NumberLimits{})
	n, err := compile.CompileDocument(s, "file:///s.json")
	if err != nil {
		t.Fatal(err)
	}
	d, _ := jsonvalue.DecodeLimited([]byte(`{"p":"x"}`), jsonvalue.NumberLimits{})
	_, units, err := eval.EvaluateUnits(n, d, eval.DefaultMaxOutputUnits)
	if err != nil || len(units) == 0 || units[0].KeywordLocation != "/properties/p/$ref/minLength" {
		t.Fatalf("%+v %v", units, err)
	}
	d, _ = jsonvalue.DecodeLimited([]byte(`"x"`), jsonvalue.NumberLimits{})
	_, units, err = eval.EvaluateUnits(n, d, eval.DefaultMaxOutputUnits)
	if err != nil || len(units) == 0 || units[0].KeywordLocation != "/then/$ref/minLength" {
		t.Fatalf("%+v %v", units, err)
	}
}

// Every schema-node entry, including through $ref, counts one evaluation
// step: $ref branches that double at each level (2^64 entries) end with the
// "evaluation steps" LimitError instead of running forever.
func TestExponentialRefBranchingHitsStepLimit(t *testing.T) {
	const levels = 64
	var b strings.Builder
	b.WriteString(`{"$ref":"#/$defs/d0","$defs":{`)
	for i := 0; i < levels; i++ {
		fmt.Fprintf(&b, `"d%d":{"allOf":[{"$ref":"#/$defs/d%d"},{"$ref":"#/$defs/d%d"}]},`, i, i+1, i+1)
	}
	fmt.Fprintf(&b, `"d%d":{"type":"integer"}}}`, levels)
	s, err := jsonvalue.DecodeLimited([]byte(b.String()), jsonvalue.NumberLimits{})
	if err != nil {
		t.Fatal(err)
	}
	n, err := compile.CompileDocument(s, "file:///s.json")
	if err != nil {
		t.Fatal(err)
	}
	d, _ := jsonvalue.DecodeLimited([]byte(`1`), jsonvalue.NumberLimits{})
	const max = 1000
	ok, _, err := eval.EvaluateLimited(n, d, eval.DefaultMaxOutputUnits, max)
	var le *jsonvalue.LimitError
	if ok || !errors.As(err, &le) || le.Limit != "evaluation steps" || le.Max != max {
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
