package applicator

import "github.com/neoface/jsonschema/internal/keywords"

func init() {
	keywords.Register(keywords.Keyword{Name: "then", Vocabulary: VocabularyURI,
		Compile:  func(ctx *keywords.CompileContext, v any) (any, error) { return compileBranch("then", ctx, v) },
		Evaluate: evalBranch})
}
