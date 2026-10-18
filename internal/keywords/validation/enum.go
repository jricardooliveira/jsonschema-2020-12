package validation

import (
	"errors"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "enum", Vocabulary: VocabularyURI, Compile: compileEnum, Evaluate: evalEnum})
}

// sized is an immutable compiled value with its canonical size in bytes,
// computed once at compile time for step charging.
type sized struct {
	v    any
	size int
}

// compileEnum requires an array; its items may be any JSON value.
func compileEnum(_ *keywords.CompileContext, value any) (any, error) {
	a, ok := value.([]any)
	if !ok {
		return nil, errors.New("enum: must be an array")
	}
	out := make([]sized, len(a))
	for i, v := range a {
		out[i] = sized{v, canonicalSize(v)}
	}
	return out, nil
}

// evalEnum charges each candidate's size before comparing: Equal stops at
// the first difference, so its cost is bounded by the candidate.
func evalEnum(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	for _, c := range compiled.([]sized) {
		if err := ctx.ChargeBytes(c.size); err != nil {
			return false, err
		}
		if jsonvalue.Equal(c.v, instance) {
			return true, nil
		}
	}
	return false, ctx.Report("value must equal one of the values in enum")
}
