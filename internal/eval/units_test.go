package eval

import (
	"errors"
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
	_ "github.com/neoface/jsonschema/internal/keywords/all"
)

// x-test-passing is a test-only applicator: it evaluates every child, discards
// their units and always passes (like a passing anyOf).
// x-test-keep evaluates one child and keeps its units before its own.
func init() {
	comp := func(_ *keywords.CompileContext, v any) (any, error) {
		var nodes []*compile.Node
		for _, e := range v.([]any) {
			n, err := compile.Compile(e)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, n)
		}
		return nodes, nil
	}
	keywords.Register(keywords.Keyword{Name: "x-test-passing", Vocabulary: "test", Compile: comp,
		Evaluate: func(ctx *keywords.EvalContext, c any, inst any) (bool, error) {
			for _, n := range c.([]*compile.Node) {
				if _, _, err := ctx.EvalChild(n, inst, ctx.InstanceLocation); err != nil {
					return false, err
				}
			}
			return true, nil
		}})
	keywords.Register(keywords.Keyword{Name: "x-test-keep", Vocabulary: "test", Compile: comp,
		Evaluate: func(ctx *keywords.EvalContext, c any, inst any) (bool, error) {
			ok, units, err := ctx.EvalChild(c.([]*compile.Node)[0], inst, ctx.InstanceLocation+"/a")
			if err != nil || ok {
				return ok, err
			}
			ctx.Keep(units)
			return false, ctx.Report("child failed")
		}})
}

func mustCompile(t *testing.T, v any) *compile.Node {
	t.Helper()
	n, err := compile.Compile(v)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestDiscardedUnitsCountAgainstCap(t *testing.T) {
	failing := map[string]any{"type": "string"}
	kids := make([]any, 10)
	for i := range kids {
		kids[i] = failing
	}
	n := mustCompile(t, map[string]any{"x-test-passing": kids})
	doc, err := jsonvalue.Decode([]byte(`1`))
	if err != nil {
		t.Fatal(err)
	}
	ok, units, err := EvaluateUnits(n, doc, DefaultMaxOutputUnits)
	if err != nil || !ok || len(units) != 0 {
		t.Fatalf("ok=%v units=%v err=%v", ok, units, err)
	}
	_, _, err = EvaluateUnits(n, doc, len(kids)-1)
	var le *jsonvalue.LimitError
	if !errors.As(err, &le) || le.Limit != "output units" {
		t.Fatalf("want output units LimitError, got %v", err)
	}
}

func TestUnitOrderAndNoValues(t *testing.T) {
	const secret = "s3cr3t-value"
	n := mustCompile(t, map[string]any{"x-test-keep": []any{map[string]any{"const": false, "type": "number"}}})
	ok, units, err := EvaluateUnits(n, secret, DefaultMaxOutputUnits)
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if len(units) != 3 {
		t.Fatalf("units=%v", units)
	}
	// child units (sorted keyword order: const, type) come before the wrapper
	if units[0].KeywordLocation != "/const" || units[1].KeywordLocation != "/type" ||
		units[2].KeywordLocation != "/x-test-keep" || units[0].InstanceLocation != "/a" || units[2].InstanceLocation != "" {
		t.Errorf("order: %+v", units)
	}
	for _, u := range units {
		if strings.Contains(u.Error+u.InstanceLocation+u.KeywordLocation, secret) {
			t.Errorf("value leaked: %+v", u)
		}
	}
}
