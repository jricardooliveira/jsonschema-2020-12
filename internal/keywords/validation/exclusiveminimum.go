package validation

import (
	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "exclusiveMinimum", Vocabulary: VocabularyURI, Compile: compileExclusiveMinimum, Evaluate: evalExclusiveMinimum})
}

func compileExclusiveMinimum(_ *keywords.CompileContext, value any) (any, error) {
	return numberValue("exclusiveMinimum", value)
}

func evalExclusiveMinimum(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	n, ok := numberInstance(instance)
	if !ok || jsonvalue.Cmp(n, compiled.(*jsonvalue.Number)) > 0 {
		return true, nil
	}
	return false, ctx.Report("value is not greater than exclusiveMinimum")
}
