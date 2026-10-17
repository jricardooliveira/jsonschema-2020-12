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

// CountStep counts one evaluation step (a member visited, a pattern match
// attempt or a reported failure) in the per-call counter held by ctx.Items.
// A *LimitError past the cap must be returned as is, never swallowed.
func CountStep(ctx *keywords.EvalContext) error {
	if s, ok := ctx.Items.(interface{ Step() error }); ok {
		return s.Step()
	}
	return nil
}

// reportMissing reports each name absent from obj, in schema order. The
// messages hold only names from the schema. Each name is charged by its
// length before the lookup, and each failure counts one more step.
func reportMissing(ctx *keywords.EvalContext, names []string, obj map[string]any) (bool, error) {
	valid := true
	for _, n := range names {
		if err := ctx.ChargeBytes(len(n)); err != nil {
			return false, err
		}
		if _, ok := obj[n]; !ok {
			valid = false
			if err := CountStep(ctx); err != nil {
				return false, err
			}
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
