package validation

import (
	"unicode/utf8"

	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "minLength", Vocabulary: VocabularyURI, Compile: compileMinLength, Evaluate: evalMinLength})
}

func compileMinLength(_ *keywords.CompileContext, value any) (any, error) {
	return compileLength("minLength", value)
}

func evalMinLength(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	s, ok := instance.(string)
	if !ok || jsonvalue.Cmp(number(utf8.RuneCountInString(s)), compiled.(*jsonvalue.Number)) >= 0 {
		return true, nil
	}
	return false, ctx.Report("string is shorter than minLength")
}
