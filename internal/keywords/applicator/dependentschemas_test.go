package applicator_test

import (
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/eval"
)

// The testdata/applicator/dependentschemas.json cases run through the root TestSpecCases.

func TestDependentSchemasAppliesOnlyWhenTriggerPresent(t *testing.T) {
	n := mustCompile(t, `{"dependentSchemas":{"a/b":{"type":"array"}}}`)
	ok, units, err := eval.EvaluateUnits(n, mustDecode(t, `{"other":1}`), 100)
	if err != nil || !ok || len(units) != 0 {
		t.Fatalf("absent trigger: ok=%v err=%v units=%+v", ok, err, units)
	}
	ok, units, err = eval.EvaluateUnits(n, mustDecode(t, `{"a/b":1}`), 100)
	if err != nil || ok || len(units) == 0 {
		t.Fatalf("present trigger: ok=%v err=%v units=%+v", ok, err, units)
	}
	if !strings.HasPrefix(units[0].KeywordLocation, "/dependentSchemas/a~1b") {
		t.Errorf("keyword location %q", units[0].KeywordLocation)
	}
}

func TestDependentSchemasSchemaError(t *testing.T) {
	for _, s := range []string{`{"dependentSchemas":[]}`, `{"dependentSchemas":{"a":"x"}}`} {
		_, err := compile.Compile(mustDecode(t, s))
		if err == nil || !strings.HasPrefix(err.Error(), "dependentSchemas: ") {
			t.Errorf("%s: err=%v", s, err)
		}
	}
}
