package validation

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "maxLength", Vocabulary: VocabularyURI, Compile: compileMaxLength, Evaluate: evalMaxLength})
}

// number turns a count into an exact Number so it can be compared without
// any overflowing conversion.
func number(n int) *jsonvalue.Number {
	v, err := jsonvalue.DecodeLimited([]byte(strconv.Itoa(n)), jsonvalue.NumberLimits{})
	if err != nil {
		panic(err) // unreachable: a small decimal integer always decodes
	}
	return v.(*jsonvalue.Number)
}

// compileLength accepts a non-negative integer (2.0 counts as 2), any size.
func compileLength(name string, value any) (any, error) {
	n, ok := value.(*jsonvalue.Number)
	if !ok || !n.IsInteger() || jsonvalue.Cmp(n, number(0)) < 0 {
		return nil, fmt.Errorf("%s: must be a non-negative integer", name)
	}
	return n, nil
}

func compileMaxLength(_ *keywords.CompileContext, value any) (any, error) {
	return compileLength("maxLength", value)
}

// runeCount charges the work of counting s's code points, then counts them.
func runeCount(ctx *keywords.EvalContext, s string) (int, error) {
	if err := ctx.ChargeBytes(len(s)); err != nil {
		return 0, err
	}
	return utf8.RuneCountInString(s), nil
}

func evalMaxLength(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	s, ok := instance.(string)
	if !ok {
		return true, nil
	}
	n, err := runeCount(ctx, s)
	if err != nil {
		return false, err
	}
	if jsonvalue.Cmp(number(n), compiled.(*jsonvalue.Number)) <= 0 {
		return true, nil
	}
	return false, ctx.Report("string is longer than maxLength")
}
