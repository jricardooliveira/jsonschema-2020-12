package applicator

import "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"

func init() {
	keywords.Register(keywords.Keyword{Name: "else", Vocabulary: VocabularyURI,
		Compile:  func(ctx *keywords.CompileContext, v any) (any, error) { return compileBranch("else", ctx, v) },
		Evaluate: evalBranch})
}
