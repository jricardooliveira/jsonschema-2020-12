package validation

import (
	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "maxProperties", Vocabulary: VocabularyURI, Compile: compileCount("maxProperties"), Evaluate: evalMaxProperties})
}

// evalMaxProperties: member count must be <= the limit (validation 6.5.1).
func evalMaxProperties(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	obj, ok := instance.(map[string]any)
	if !ok {
		return true, nil
	}
	c, err := cmpCount(len(obj), compiled.(*jsonvalue.Number))
	if err != nil {
		return false, err
	}
	if c <= 0 {
		return true, nil
	}
	return false, ctx.Report("object has more properties than maxProperties allows")
}
