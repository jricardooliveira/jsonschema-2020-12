package core

import (
	"fmt"

	"github.com/neoface/jsonschema/internal/keywords"
)

// VocabularyURI is the 2020-12 core vocabulary.
const VocabularyURI = "https://json-schema.org/draft/2020-12/vocab/core"

func init() {
	keywords.Register(keywords.Keyword{Name: "$id", Vocabulary: VocabularyURI, Compile: compileID, Evaluate: noEffect})
}

// noEffect is the evaluator of $id, $anchor, $defs (core 8.2.4) and $comment
// (core 8.3): they never change a result. Base URIs are handled at load time.
func noEffect(*keywords.EvalContext, any, any) (bool, error) { return true, nil }

func compileID(_ *keywords.CompileContext, value any) (any, error) {
	if _, ok := value.(string); !ok {
		return nil, fmt.Errorf("$id: must be a string")
	}
	return nil, nil
}
