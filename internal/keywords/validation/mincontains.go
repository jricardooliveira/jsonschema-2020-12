package validation

import "github.com/neoface/jsonschema/internal/keywords"

// minContains and maxContains are read by the sibling contains at compile
// time; on their own they check only their value and never fail an instance.
func init() {
	keywords.Register(keywords.Keyword{Name: "minContains", Vocabulary: VocabularyURI, Compile: compileCount("minContains"), Evaluate: passAlways})
}

func passAlways(*keywords.EvalContext, any, any) (bool, error) { return true, nil }
