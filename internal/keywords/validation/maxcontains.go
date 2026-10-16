package validation

import "github.com/neoface/jsonschema/internal/keywords"

func init() {
	keywords.Register(keywords.Keyword{Name: "maxContains", Vocabulary: VocabularyURI, Compile: compileCount("maxContains"), Evaluate: passAlways})
}
