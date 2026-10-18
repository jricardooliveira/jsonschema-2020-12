package applicator

import "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"

func init() {
	keywords.Register(keywords.Keyword{Name: "then", Vocabulary: VocabularyURI,
		Compile:  func(ctx *keywords.CompileContext, v any) (any, error) { return compileBranch("then", ctx, v) },
		Evaluate: evalBranch})
}
