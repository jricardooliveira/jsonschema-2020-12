package core

import "github.com/neoface/jsonschema/internal/keywords"

func init() {
	keywords.Register(keywords.Keyword{Name: "$schema", Vocabulary: VocabularyURI, Compile: compileAny, Evaluate: noEffect})
}

// compileAny accepts any value: the meta-schema check (core 8.1.1) reports a
// wrong '$schema' or '$vocabulary' type, so compiling must not.
func compileAny(*keywords.CompileContext, any) (any, error) { return nil, nil }
