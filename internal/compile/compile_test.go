package compile

import (
	"errors"
	"testing"

	"github.com/neoface/jsonschema/internal/keywords"
)

func TestCompileShapes(t *testing.T) {
	for _, v := range []any{true, false, map[string]any{}, map[string]any{"foo": 1}} {
		if _, err := Compile(v); err != nil {
			t.Errorf("%v: %v", v, err)
		}
	}
	for _, v := range []any{nil, "x", []any{}, 5} {
		if _, err := Compile(v); !errors.Is(err, ErrNotSchema) {
			t.Errorf("%v: got %v", v, err)
		}
	}
}

func TestRegistryKeywordReached(t *testing.T) {
	calls := 0
	keywords.Register(keywords.Keyword{
		Name: "x-compile-test-kw", Vocabulary: "urn:test",
		Compile: func(_ *keywords.CompileContext, v any) (any, error) { return v, nil },
		Evaluate: func(_ *keywords.EvalContext, c any, _ any) (bool, error) {
			calls++
			return c.(bool), nil
		},
	})
	n, err := Compile(map[string]any{"x-compile-test-kw": false})
	if err != nil || len(n.Items) != 1 {
		t.Fatalf("compile: %v %v", n, err)
	}
	ok, err := n.Items[0].Keyword.Evaluate(&keywords.EvalContext{}, n.Items[0].Data, nil)
	if ok || err != nil || calls != 1 {
		t.Fatalf("evaluate: %v %v %d", ok, err, calls)
	}
}
