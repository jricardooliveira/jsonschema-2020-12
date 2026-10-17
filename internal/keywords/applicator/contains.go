package applicator

import (
	"fmt"
	"strconv"

	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "contains", Vocabulary: VocabularyURI, Compile: compileContains, Evaluate: evalContains})
}

// compiledContains is built once at compile time and only read afterwards.
type compiledContains struct {
	node     any
	min, max *jsonvalue.Number // nil when the sibling keyword is absent
}

func compileContains(ctx *keywords.CompileContext, value any) (any, error) {
	n, err := ctx.Sub(value, "")
	if err != nil {
		return nil, fmt.Errorf("contains: %w", err)
	}
	c := &compiledContains{node: n}
	if v, ok := ctx.Siblings["minContains"].(*jsonvalue.Number); ok {
		c.min = v
	}
	if v, ok := ctx.Siblings["maxContains"].(*jsonvalue.Number); ok {
		c.max = v
	}
	return c, nil
}

// cmpCount compares a match count with a keyword value exactly.
func cmpCount(count int, limit *jsonvalue.Number) (int, error) {
	c, err := jsonvalue.DecodeLimited([]byte(strconv.Itoa(count)), jsonvalue.NumberLimits{})
	if err != nil {
		return 0, err
	}
	return jsonvalue.Cmp(c.(*jsonvalue.Number), limit), nil
}

func evalContains(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	arr, ok := instance.([]any)
	if !ok {
		return true, nil
	}
	c := compiled.(*compiledContains)
	var matches []int
	for i, el := range arr {
		if err := ctx.Charge(1); err != nil {
			return false, err
		}
		// Non-matching elements' units are dropped; EvalChild already counted them.
		ok, _, err := ctx.EvalChild(c.node, el, ctx.InstanceLocation+"/"+strconv.Itoa(i))
		if err != nil {
			return false, err
		}
		if ok {
			matches = append(matches, i)
		}
	}
	n := strconv.Itoa(len(matches))
	if c.min != nil {
		cmp, err := cmpCount(len(matches), c.min)
		if err != nil {
			return false, err
		}
		if cmp < 0 {
			return false, ctx.Report("minContains: " + n + " elements match contains, at least " + c.min.String() + " required")
		}
	} else if len(matches) == 0 {
		return false, ctx.Report("contains: no element matches, at least 1 required")
	}
	if c.max != nil {
		cmp, err := cmpCount(len(matches), c.max)
		if err != nil {
			return false, err
		}
		if cmp > 0 {
			return false, ctx.Report("maxContains: " + n + " elements match contains, at most " + c.max.String() + " allowed")
		}
	}
	if len(matches) > 0 {
		if rec, ok := ctx.Items.(keywords.ItemRecorder); ok {
			rec.RecordContains(ctx.InstanceLocation, ctx.KeywordLoc(), matches, len(matches) == len(arr))
		}
		var v any = matches
		if len(matches) == len(arr) {
			v = true
		}
		return true, ctx.Annotate(v)
	}
	return true, nil
}
