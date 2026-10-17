package keywords

import (
	"errors"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
)

func TestAnnotateOnlyAnnotatesAndPasses(t *testing.T) {
	k := AnnotateOnly("x", "v")
	data, err := k.Compile(&CompileContext{}, map[string]any{"a": "b"})
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewEvalContext(10, nil)
	ctx.Keyword = "x"
	ok, err := k.Evaluate(ctx, data, "instance")
	if !ok || err != nil {
		t.Fatal(ok, err)
	}
	a := ctx.Annotations()
	if len(a) != 1 || a[0].Keyword != "x" || a[0].Value == nil {
		t.Fatalf("got %+v", a)
	}
	if len(ctx.Units()) != 0 {
		t.Fatal("must not report failures")
	}
}

func TestAnnotateOnlyOverLimitIsError(t *testing.T) {
	k := AnnotateOnly("x", "v")
	ctx := NewEvalContext(10, nil)
	ctx.SetMaxAnnotations(1)
	if _, err := k.Evaluate(ctx, 1, nil); err != nil {
		t.Fatal(err)
	}
	_, err := k.Evaluate(ctx, 1, nil)
	var le *jsonvalue.LimitError
	if !errors.As(err, &le) {
		t.Fatalf("want LimitError, got %v", err)
	}
}
