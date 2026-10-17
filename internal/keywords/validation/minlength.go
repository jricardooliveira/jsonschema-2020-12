package validation

import (
	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "minLength", Vocabulary: VocabularyURI, Compile: compileMinLength, Evaluate: evalMinLength})
}

func compileMinLength(_ *keywords.CompileContext, value any) (any, error) {
	return compileLength("minLength", value)
}

func evalMinLength(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	s, ok := instance.(string)
	if !ok {
		return true, nil
	}
	n, err := runeCount(ctx, s)
	if err != nil {
		return false, err
	}
	if jsonvalue.Cmp(number(n), compiled.(*jsonvalue.Number)) >= 0 {
		return true, nil
	}
	return false, ctx.Report("string is shorter than minLength")
}
