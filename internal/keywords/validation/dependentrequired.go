package validation

import (
	"errors"
	"sort"

	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "dependentRequired", Vocabulary: VocabularyURI, Compile: compileDependentRequired, Evaluate: evalDependentRequired})
}

type depRequired struct {
	trigger string
	names   []string
}

func compileDependentRequired(_ *keywords.CompileContext, value any) (any, error) {
	m, ok := value.(map[string]any)
	if !ok {
		return nil, errors.New("dependentRequired: must be an object")
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]depRequired, 0, len(keys))
	for _, k := range keys {
		names, err := compileNames("dependentRequired", m[k])
		if err != nil {
			return nil, err
		}
		out = append(out, depRequired{k, names})
	}
	return out, nil
}

// evalDependentRequired: triggers walk in sorted order, missing names in
// schema order (validation 6.5.4).
func evalDependentRequired(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	obj, ok := instance.(map[string]any)
	if !ok {
		return true, nil
	}
	valid := true
	for _, d := range compiled.([]depRequired) {
		if err := ctx.ChargeBytes(len(d.trigger)); err != nil { // per entry, by key length

			return false, err
		}
		if _, present := obj[d.trigger]; !present {
			continue
		}
		v, err := reportMissing(ctx, d.names, obj)
		if err != nil {
			return false, err
		}
		valid = valid && v
	}
	return valid, nil
}
