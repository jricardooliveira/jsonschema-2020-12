package validation

import (
	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "const", Vocabulary: VocabularyURI, Compile: compileConst, Evaluate: evalConst})
}

// compileConst accepts any JSON value and keeps its canonical size.
func compileConst(_ *keywords.CompileContext, value any) (any, error) {
	return sized{value, canonicalSize(value)}, nil
}

func evalConst(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	c := compiled.(sized)
	if err := ctx.ChargeBytes(c.size); err != nil {
		return false, err
	}
	if jsonvalue.Equal(c.v, instance) {
		return true, nil
	}
	return false, ctx.Report("value must equal the const value")
}
