package validation

import (
	"errors"
	"strconv"

	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "maxItems", Vocabulary: VocabularyURI, Compile: compileMaxItems, Evaluate: evalMaxItems})
}

// compileCount requires a number; the exact Number is kept so 2.0 behaves as 2
// and huge, negative or fractional values are compared, never converted to int.
func compileCount(name string) keywords.CompileFunc {
	return func(_ *keywords.CompileContext, value any) (any, error) {
		n, ok := value.(*jsonvalue.Number)
		if !ok {
			return nil, errors.New(name + ": must be a number")
		}
		return n, nil
	}
}

// cmpCount compares an element count with a keyword value exactly.
func cmpCount(count int, limit *jsonvalue.Number) (int, error) {
	c, err := jsonvalue.DecodeLimited([]byte(strconv.Itoa(count)), jsonvalue.NumberLimits{})
	if err != nil {
		return 0, err
	}
	return jsonvalue.Cmp(c.(*jsonvalue.Number), limit), nil
}

var compileMaxItems = compileCount("maxItems")

func evalMaxItems(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	arr, ok := instance.([]any)
	if !ok {
		return true, nil
	}
	c, err := cmpCount(len(arr), compiled.(*jsonvalue.Number))
	if err != nil {
		return false, err
	}
	if c <= 0 {
		return true, nil
	}
	return false, ctx.Report("array has more elements than maxItems allows (" + strconv.Itoa(len(arr)) + " found)")
}
