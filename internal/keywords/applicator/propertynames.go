package applicator

import (
	"fmt"
	"sort"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "propertyNames", Vocabulary: VocabularyURI, Compile: compilePropertyNames, Evaluate: evalPropertyNames})
}

func compilePropertyNames(ctx *keywords.CompileContext, value any) (any, error) {
	n, err := ctx.Sub(value, "")
	if err != nil {
		return nil, fmt.Errorf("propertyNames: %w", err)
	}
	return n, nil
}

// evalPropertyNames validates each member name as a string instance (core
// 10.3.2.4). The name is evaluated at the object's own location, so it never
// appears in a location or message. Names are visited in sorted order.
func evalPropertyNames(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	obj, ok := instance.(map[string]any)
	if !ok {
		return true, nil
	}
	names := make([]string, 0, len(obj))
	for k := range obj {
		names = append(names, k)
	}
	sort.Strings(names)
	valid := true
	for _, k := range names {
		if err := step(ctx); err != nil {
			return false, err
		}
		ok, units, err := ctx.EvalChild(compiled, k, ctx.InstanceLocation)
		if err != nil {
			return false, err
		}
		if !ok {
			valid = false
			ctx.Keep(units)
			if err := step(ctx); err != nil {
				return false, err
			}
		}
	}
	return valid, nil
}
