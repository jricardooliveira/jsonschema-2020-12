// Package eval runs a compiled schema against a decoded document.
package eval

import (
	"strings"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/keywords"
)

// Evaluate reports whether instance is valid. An error is never valid.
// It only reads n; per-call state lives in the EvalContext.
func Evaluate(n *compile.Node, instance any) (bool, error) {
	ok, _, err := EvaluateUnits(n, instance, DefaultMaxOutputUnits)
	return ok, err
}

// evalNode evaluates a compiled schema into ctx. Units of failing keywords
// and applicators accumulate in ctx in evaluation order.
func evalNode(ctx *keywords.EvalContext, node any, instance any) (bool, error) {
	n := node.(*compile.Node)
	attachItems(ctx)
	pf, leave, err := enterNode(ctx, n)
	if err != nil {
		return false, err
	}
	defer leave()
	if ctx.Tree != nil {
		ctx.Tree.KeywordLocation = pf.dynamic
	}
	// core 12.3.2: absoluteKeywordLocation only once a reference was crossed.
	abs := func(loc string) string {
		if !pf.viaRef {
			return ""
		}
		return n.Resource.AbsoluteLocation(loc)
	}
	if n.Bool != nil {
		if *n.Bool {
			return true, nil
		}
		ctx.KeywordLocation = pf.dynamic
		ctx.AbsoluteKeywordLocation = abs(n.Location)
		return false, ctx.Report("schema is false, no value is valid")
	}
	valid := true
	// Two passes: keywords flagged AfterSiblings run last, in schema order.
	for pass := 0; pass < 2; pass++ {
		for _, it := range n.Items {
			if it.Keyword.AfterSiblings != (pass == 1) {
				continue
			}
			ok, err := evalItem(ctx, pf, abs, it, instance)
			if err != nil {
				return false, err
			}
			if !ok {
				valid = false
			}
		}
	}
	return valid, nil
}

// CoveredItems summarises the item annotations at instLoc in plain types, for
// keyword packages that cannot import eval (unevaluatedItems, core 11.2): all
// is true when one annotation covers every element, prefix is the largest
// index prefixItems covered (-1 when none) and indices are contains' matches.
func (r *EvaluatedItems) CoveredItems(instLoc string) (all bool, prefix int, indices []int) {
	prefix = -1
	for kw, a := range r.At(instLoc) {
		switch {
		case a.All:
			all = true
		case strings.HasSuffix(kw, "/prefixItems"):
			prefix = max(prefix, a.Largest)
		default:
			indices = append(indices, a.Indices...)
		}
	}
	return all, prefix, indices
}

// evalItem evaluates one keyword of a schema object and makes sure a failure
// leaves at least one unit.
func evalItem(ctx *keywords.EvalContext, pf *pathFrame, abs func(string) string, it compile.Item, instance any) (bool, error) {
	ctx.KeywordLocation = pf.at(it.Keyword.Name, it.Location)
	ctx.AbsoluteKeywordLocation = abs(it.Location)
	ctx.Keyword = it.Keyword.Name
	before := len(ctx.Units())
	ok, err := it.Keyword.Evaluate(ctx, it.Data, instance)
	if err != nil {
		return false, err
	}
	if !ok && len(ctx.Units()) == before { // keyword gave no unit of its own
		if err := ctx.Report("rule " + it.Keyword.Name + " failed"); err != nil {
			return false, err
		}
	}
	return ok, nil
}
