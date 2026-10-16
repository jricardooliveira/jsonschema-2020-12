package eval

import (
	"errors"
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/keywords"
)

var errTestEval = errors.New("test evaluate failure")

// Registered once per process (init), so -count=N is safe.
func init() {
	keywords.Register(keywords.Keyword{
		Name: "x-eval-test-bool", Vocabulary: "urn:test",
		Compile:  func(_ *keywords.CompileContext, v any) (any, error) { return v, nil },
		Evaluate: func(_ *keywords.EvalContext, c any, _ any) (bool, error) { return c.(bool), nil },
	})
	keywords.Register(keywords.Keyword{
		Name: "x-eval-test-err", Vocabulary: "urn:test",
		Compile:  func(*keywords.CompileContext, any) (any, error) { return nil, nil },
		Evaluate: func(*keywords.EvalContext, any, any) (bool, error) { return true, errTestEval },
	})
}

func run(t *testing.T, schema map[string]any) (bool, error) {
	t.Helper()
	n, err := compile.Compile(schema)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	return Evaluate(n, map[string]any{})
}

func TestEvaluateDispatchesRegisteredKeyword(t *testing.T) {
	for _, want := range []bool{true, false} {
		got, err := run(t, map[string]any{"x-eval-test-bool": want})
		if err != nil || got != want {
			t.Errorf("keyword returns %v: got %v %v", want, got, err)
		}
	}
}

func TestEvaluateErrorPropagates(t *testing.T) {
	ok, err := run(t, map[string]any{"x-eval-test-err": 1, "x-eval-test-bool": true})
	if ok || !errors.Is(err, errTestEval) {
		t.Fatalf("want errTestEval and not valid, got %v %v", ok, err)
	}
}
