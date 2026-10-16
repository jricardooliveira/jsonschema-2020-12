package validation

import (
	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "maximum", Vocabulary: VocabularyURI, Compile: compileMaximum, Evaluate: evalMaximum})
}

func compileMaximum(_ *keywords.CompileContext, value any) (any, error) {
	return numberValue("maximum", value)
}

func evalMaximum(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	n, ok := numberInstance(instance)
	if !ok || jsonvalue.Cmp(n, compiled.(*jsonvalue.Number)) <= 0 {
		return true, nil
	}
	return false, ctx.Report("value is greater than maximum")
}
