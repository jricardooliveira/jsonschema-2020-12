package validation

import (
	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "exclusiveMaximum", Vocabulary: VocabularyURI, Compile: compileExclusiveMaximum, Evaluate: evalExclusiveMaximum})
}

func compileExclusiveMaximum(_ *keywords.CompileContext, value any) (any, error) {
	return numberValue("exclusiveMaximum", value)
}

func evalExclusiveMaximum(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	n, ok := numberInstance(instance)
	if !ok || jsonvalue.Cmp(n, compiled.(*jsonvalue.Number)) < 0 {
		return true, nil
	}
	return false, ctx.Report("value is not less than exclusiveMaximum")
}
