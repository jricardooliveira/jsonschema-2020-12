package validation

import (
	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "minProperties", Vocabulary: VocabularyURI, Compile: compileCount("minProperties"), Evaluate: evalMinProperties})
}

// evalMinProperties: member count must be >= the limit (validation 6.5.2).
func evalMinProperties(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	obj, ok := instance.(map[string]any)
	if !ok {
		return true, nil
	}
	c, err := cmpCount(len(obj), compiled.(*jsonvalue.Number))
	if err != nil {
		return false, err
	}
	if c >= 0 {
		return true, nil
	}
	return false, ctx.Report("object has fewer properties than minProperties requires")
}
