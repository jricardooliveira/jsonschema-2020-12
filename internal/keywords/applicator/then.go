package applicator

import "github.com/neoface/jsonschema/internal/keywords"

func init() {
	keywords.Register(keywords.Keyword{Name: "then", Vocabulary: VocabularyURI,
		Compile: func(ctx *keywords.CompileContext, v any) (any, error) { return compileBranch("then", ctx, v) },
		Evaluate: func(ctx *keywords.EvalContext, c any, inst any) (bool, error) {
			return evalBranch("then", true, ctx, c, inst)
		}})
}
