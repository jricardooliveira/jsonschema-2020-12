package validation

import (
	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "minimum", Vocabulary: VocabularyURI, Compile: compileMinimum, Evaluate: evalMinimum})
}

func compileMinimum(_ *keywords.CompileContext, value any) (any, error) {
	return numberValue("minimum", value)
}

func evalMinimum(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	n, ok := numberInstance(instance)
	if !ok || jsonvalue.Cmp(n, compiled.(*jsonvalue.Number)) >= 0 {
		return true, nil
	}
	return false, ctx.Report("value is less than minimum")
}
