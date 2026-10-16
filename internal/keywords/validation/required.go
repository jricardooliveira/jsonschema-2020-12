package validation

import (
	"errors"

	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "required", Vocabulary: VocabularyURI, Compile: compileRequired, Evaluate: evalRequired})
}

// compileNames checks that v is an array of strings and returns it.
func compileNames(name string, v any) ([]string, error) {
	arr, ok := v.([]any)
	if !ok {
		return nil, errors.New(name + ": must be an array of strings")
	}
	out := make([]string, 0, len(arr))
	for _, e := range arr {
		s, ok := e.(string)
		if !ok {
			return nil, errors.New(name + ": must be an array of strings")
		}
		out = append(out, s)
	}
	return out, nil
}

func compileRequired(_ *keywords.CompileContext, value any) (any, error) {
	return compileNames("required", value)
}

// reportMissing reports each name absent from obj, in schema order. The
// messages hold only names from the schema.
func reportMissing(ctx *keywords.EvalContext, names []string, obj map[string]any) (bool, error) {
	valid := true
	for _, n := range names {
		if _, ok := obj[n]; !ok {
			valid = false
			if err := ctx.Report("missing required property: " + n); err != nil {
				return false, err
			}
		}
	}
	return valid, nil
}

// evalRequired (validation 6.5.3).
func evalRequired(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	obj, ok := instance.(map[string]any)
	if !ok {
		return true, nil
	}
	return reportMissing(ctx, compiled.([]string), obj)
}
