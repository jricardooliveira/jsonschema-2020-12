package output

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/eval"
	"github.com/neoface/jsonschema/internal/keywords"
	_ "github.com/neoface/jsonschema/internal/keywords/all"
)

// x-out-keep is a test-only applicator: it compiles its value as a child at
// /x-out-keep/0 and keeps the child's units before its own.
func init() {
	keywords.Register(keywords.Keyword{Name: "x-out-keep", Vocabulary: "test",
		Compile: func(cc *keywords.CompileContext, v any) (any, error) { return cc.Sub(v, "/0") },
		Evaluate: func(ctx *keywords.EvalContext, c any, inst any) (bool, error) {
			ok, units, err := ctx.EvalChild(c.(*compile.Node), inst, ctx.InstanceLocation)
			if err != nil || ok {
				return ok, err
			}
			ctx.Keep(units)
			return false, ctx.Report("child failed")
		}})
}

func TestBasicChildUnitsRootedAndValueFree(t *testing.T) {
	const secret = "s3cr3t-value"
	n, err := compile.Compile(map[string]any{"x-out-keep": map[string]any{"const": false}})
	if err != nil {
		t.Fatal(err)
	}
	ok, units, err := eval.EvaluateUnits(n, secret, eval.DefaultMaxOutputUnits)
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	b, err := Render("basic", ok, units)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), secret) {
		t.Fatalf("value leaked: %s", b)
	}
	var got struct {
		Errors []basicUnit `json:"errors"`
	}
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Errors) != 2 || got.Errors[0].KeywordLocation != "/x-out-keep/0/const" ||
		got.Errors[1].KeywordLocation != "/x-out-keep" {
		t.Fatalf("errors: %s", b)
	}
}
