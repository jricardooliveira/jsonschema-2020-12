package core

import (
	"fmt"

	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "$defs", Vocabulary: VocabularyURI, Compile: compileDefs, Evaluate: noEffect})
}

// compileDefs only checks the shape; $defs members are compiled when a $ref
// reaches them, so they have no effect unless referenced.
func compileDefs(_ *keywords.CompileContext, value any) (any, error) {
	if _, ok := value.(map[string]any); !ok {
		return nil, fmt.Errorf("$defs: must be an object")
	}
	return nil, nil
}
