package validation

import (
	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "const", Vocabulary: VocabularyURI, Compile: compileConst, Evaluate: evalConst})
}

// compileConst accepts any JSON value.
func compileConst(_ *keywords.CompileContext, value any) (any, error) { return value, nil }

func evalConst(_ *keywords.EvalContext, compiled any, instance any) (bool, error) {
	return jsonvalue.Equal(compiled, instance), nil
}
