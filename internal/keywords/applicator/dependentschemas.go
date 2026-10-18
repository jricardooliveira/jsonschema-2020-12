package applicator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "dependentSchemas", Vocabulary: VocabularyURI, Compile: compileDependentSchemas, Evaluate: evalDependentSchemas})
}

type depSchema struct {
	trigger string
	node    any
}

var depPointerEscape = strings.NewReplacer("~", "~0", "/", "~1")

// compileDependentSchemas builds the entries sorted by trigger name.
func compileDependentSchemas(ctx *keywords.CompileContext, value any) (any, error) {
	m, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("dependentSchemas: must be an object")
	}
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	sort.Strings(names)
	out := make([]depSchema, 0, len(names))
	for _, k := range names {
		n, err := ctx.Sub(m[k], "/"+depPointerEscape.Replace(k))
		if err != nil {
			return nil, fmt.Errorf("dependentSchemas: %w", err)
		}
		out = append(out, depSchema{k, n})
	}
	return out, nil
}

// evalDependentSchemas applies each subschema to the whole object when its
// trigger property is present (core 10.2.2.4).
func evalDependentSchemas(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	obj, ok := instance.(map[string]any)
	if !ok {
		return true, nil
	}
	valid := true
	for _, d := range compiled.([]depSchema) {
		if err := ctx.ChargeBytes(len(d.trigger)); err != nil { // per entry, by key length
			return false, err
		}
		if _, present := obj[d.trigger]; !present {
			continue
		}
		ok, units, err := ctx.EvalChild(d.node, instance, ctx.InstanceLocation)
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
