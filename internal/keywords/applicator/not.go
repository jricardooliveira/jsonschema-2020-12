package applicator

import (
	"fmt"

	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "not", Vocabulary: VocabularyURI, Compile: compileNot, Evaluate: evalNot})
}

func compileNot(ctx *keywords.CompileContext, value any) (any, error) {
	n, err := ctx.Sub(value, "")
	if err != nil {
		return nil, fmt.Errorf("not: %w", err)
	}
	return n, nil
}

// evalNot (core 10.2.1.4) is valid only when the subschema fails. The child's
// units are dropped (still counted); a child error propagates.
func evalNot(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	ok, _, err := ctx.EvalChild(compiled, instance, ctx.InstanceLocation)
	if err != nil {
		return false, err
	}
	if ok {
		return false, ctx.Report("not: document matched a schema it must not match")
	}
	return true, nil
}
