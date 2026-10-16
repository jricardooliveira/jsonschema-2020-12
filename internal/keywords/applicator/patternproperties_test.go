package applicator_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/eval"
	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords/validation"
)

// Many members x many patterns must hit the evaluation step limit as an
// error, never as a valid or invalid result.
func TestPatternPropertiesStepLimit(t *testing.T) {
	const n = 50
	var schema, doc strings.Builder
	schema.WriteString(`{"patternProperties":{`)
	doc.WriteString(`{`)
	for i := 0; i < n; i++ {
		if i > 0 {
			schema.WriteString(",")
			doc.WriteString(",")
		}
		fmt.Fprintf(&schema, `"^p%d$":true`, i)
		fmt.Fprintf(&doc, `"m%d":1`, i)
	}
	schema.WriteString(`},"additionalProperties":true}`)
	doc.WriteString(`}`)
	node, inst := mustCompile(t, schema.String()), mustDecode(t, doc.String())

	_, _, err := eval.EvaluateLimited(node, inst, eval.DefaultMaxOutputUnits, n*n)
	var le *jsonvalue.LimitError
	if !errors.As(err, &le) || le.Limit != "evaluation steps" {
		t.Fatalf("want evaluation steps LimitError, got %v", err)
	}
	ok, _, err := eval.EvaluateLimited(node, inst, eval.DefaultMaxOutputUnits, eval.DefaultMaxEvalSteps)
	if err != nil || !ok {
		t.Fatalf("default limit: ok=%v err=%v", ok, err)
	}
}

// The testdata/applicator/patternproperties.json cases run through the root TestSpecCases.

func TestPatternPropertiesSchemaErrors(t *testing.T) {
	big := strings.Repeat("a", validation.DefaultMaxPatternSize+1)
	for _, s := range []string{
		`{"patternProperties":{"(?=a)":true}}`,
		`{"patternProperties":{"(":true}}`,
		`{"patternProperties":{"` + big + `":true}}`,
		`{"patternProperties":[]}`,
		`{"patternProperties":{"a":"x"}}`,
		`{"additionalProperties":true,"patternProperties":{"(?=a)":true}}`,
	} {
		if _, err := compile.Compile(mustDecode(t, s)); err == nil {
			t.Errorf("%.60s: want schema error", s)
		}
	}
}
