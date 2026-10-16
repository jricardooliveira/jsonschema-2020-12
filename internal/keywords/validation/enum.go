package validation

import (
	"errors"

	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "enum", Vocabulary: VocabularyURI, Compile: compileEnum, Evaluate: evalEnum})
}

// compileEnum requires an array; its items may be any JSON value.
func compileEnum(_ *keywords.CompileContext, value any) (any, error) {
	a, ok := value.([]any)
	if !ok {
		return nil, errors.New("enum: must be an array")
	}
	return a, nil
}

func evalEnum(_ *keywords.EvalContext, compiled any, instance any) (bool, error) {
	for _, v := range compiled.([]any) {
		if jsonvalue.Equal(v, instance) {
			return true, nil
		}
	}
	return false, nil
}
