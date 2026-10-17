package applicator

import "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"

func init() {
	keywords.Register(keywords.Keyword{Name: "allOf", Vocabulary: VocabularyURI, Compile: compileAllOf, Evaluate: evalAllOf})
}

func compileAllOf(ctx *keywords.CompileContext, value any) (any, error) {
	return compileSchemaArray("allOf", ctx, value)
}

// evalAllOf (core 10.2.1.1) checks every subschema, keeps each failing
// subschema's units, then adds its own unit.
func evalAllOf(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	valid := true
	for _, n := range compiled.([]any) {
		ok, units, err := ctx.EvalChild(n, instance, ctx.InstanceLocation)
		if err != nil {
			return false, err
		}
		if !ok {
			valid = false
			ctx.Keep(units)
		}
	}
	if !valid {
		return false, ctx.Report("allOf: not every subschema matched")
	}
	return true, nil
}
