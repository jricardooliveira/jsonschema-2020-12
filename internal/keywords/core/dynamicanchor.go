package core

import (
	"fmt"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "$dynamicAnchor", Vocabulary: VocabularyURI, Compile: compileDynamicAnchor, Evaluate: noEffect})
}

// compileDynamicAnchor keeps the anchor name as the keyword's data; the
// evaluator's dynamic-scope lookup matches on it (core 8.2.2). The name is
// indexed, and duplicates are rejected, by the registry at load time.
func compileDynamicAnchor(_ *keywords.CompileContext, value any) (any, error) {
	s, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("$dynamicAnchor: must be a string")
	}
	return s, nil
}
