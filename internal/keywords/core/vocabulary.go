package core

import "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"

func init() {
	keywords.Register(keywords.Keyword{Name: "$vocabulary", Vocabulary: VocabularyURI, Compile: compileAny, Evaluate: noEffect})
}
