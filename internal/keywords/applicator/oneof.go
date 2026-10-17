package applicator

import (
	"strconv"
	"strings"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "oneOf", Vocabulary: VocabularyURI, Compile: compileOneOf, Evaluate: evalOneOf})
}

func compileOneOf(ctx *keywords.CompileContext, value any) (any, error) {
	return compileSchemaArray("oneOf", ctx, value)
}

// evalOneOf (core 10.2.1.3) evaluates every choice. Zero matches reports like
// anyOf; two or more report one unit naming the matching indices.
func evalOneOf(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	var matches []string
	var failed []keywords.Unit
	for i, n := range compiled.([]any) {
		ok, units, err := ctx.EvalChild(n, instance, ctx.InstanceLocation)
		if err != nil {
			return false, err
		}
		if ok {
			matches = append(matches, strconv.Itoa(i))
		} else {
			failed = append(failed, units...)
		}
	}
	switch len(matches) {
	case 1:
		return true, nil
	case 0:
		ctx.Keep(failed)
		return false, ctx.Report("oneOf: no subschema matched")
	case 2:
		return false, ctx.Report("oneOf: choices " + matches[0] + " and " + matches[1] + " both matched")
	}
	last := len(matches) - 1
	return false, ctx.Report("oneOf: choices " + strings.Join(matches[:last], ", ") + " and " + matches[last] + " all matched")
}
