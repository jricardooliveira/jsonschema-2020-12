package core

import (
	"fmt"

	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "$comment", Vocabulary: VocabularyURI, Compile: compileComment, Evaluate: noEffect})
}

func compileComment(_ *keywords.CompileContext, value any) (any, error) {
	if _, ok := value.(string); !ok {
		return nil, fmt.Errorf("$comment: must be a string")
	}
	return nil, nil
}
