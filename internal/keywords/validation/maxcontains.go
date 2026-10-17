package validation

import "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"

func init() {
	keywords.Register(keywords.Keyword{Name: "maxContains", Vocabulary: VocabularyURI, Compile: compileCount("maxContains"), Evaluate: passAlways})
}
