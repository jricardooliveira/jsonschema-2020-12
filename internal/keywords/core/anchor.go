package core

import (
	"fmt"

	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "$anchor", Vocabulary: VocabularyURI, Compile: compileAnchor, Evaluate: noEffect})
}

func compileAnchor(_ *keywords.CompileContext, value any) (any, error) {
	if _, ok := value.(string); !ok {
		return nil, fmt.Errorf("$anchor: must be a string")
	}
	return nil, nil
}
