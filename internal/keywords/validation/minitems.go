package validation

import (
	"strconv"

	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "minItems", Vocabulary: VocabularyURI, Compile: compileCount("minItems"), Evaluate: evalMinItems})
}

func evalMinItems(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	arr, ok := instance.([]any)
	if !ok {
		return true, nil
	}
	c, err := cmpCount(len(arr), compiled.(*jsonvalue.Number))
	if err != nil {
		return false, err
	}
	if c >= 0 {
		return true, nil
	}
	return false, ctx.Report("array has fewer elements than minItems requires (" + strconv.Itoa(len(arr)) + " found)")
}
