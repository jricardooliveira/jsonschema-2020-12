package applicator

import "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"

func init() {
	keywords.Register(keywords.Keyword{Name: "anyOf", Vocabulary: VocabularyURI, Compile: compileAnyOf, Evaluate: evalAnyOf})
}

func compileAnyOf(ctx *keywords.CompileContext, value any) (any, error) {
	return compileSchemaArray("anyOf", ctx, value)
}

// evalAnyOf (core 10.2.1.2) evaluates every choice, never stopping at the
// first match. When none match, it keeps each choice's units, then its own.
func evalAnyOf(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	matched := false
	var failed []keywords.Unit
	for _, n := range compiled.([]any) {
		ok, units, err := ctx.EvalChild(n, instance, ctx.InstanceLocation)
		if err != nil {
			return false, err
		}
		if ok {
			matched = true
		} else {
			failed = append(failed, units...)
		}
	}
	if matched {
		return true, nil
	}
	ctx.Keep(failed)
	return false, ctx.Report("anyOf: no subschema matched")
}
