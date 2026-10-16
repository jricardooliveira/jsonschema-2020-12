package applicator

import "github.com/neoface/jsonschema/internal/keywords"

func init() {
	keywords.Register(keywords.Keyword{Name: "else", Vocabulary: VocabularyURI,
		Compile: func(ctx *keywords.CompileContext, v any) (any, error) { return compileBranch("else", ctx, v) },
		Evaluate: func(ctx *keywords.EvalContext, c any, inst any) (bool, error) {
			return evalBranch("else", false, ctx, c, inst)
		}})
}
