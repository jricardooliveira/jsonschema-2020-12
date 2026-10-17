package core

import "github.com/neoface/jsonschema/internal/keywords"

func init() {
	keywords.Register(keywords.Keyword{Name: "$vocabulary", Vocabulary: VocabularyURI, Compile: compileAny, Evaluate: noEffect})
}
