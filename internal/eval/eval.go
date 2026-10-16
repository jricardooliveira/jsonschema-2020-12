// Package eval runs a compiled schema against a decoded document.
package eval

import (
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
	if n.Bool != nil {
		if *n.Bool {
			return true, nil
		}
		ctx.KeywordLocation = n.Location
		return false, ctx.Report("schema is false, no value is valid")
	}
	valid := true
	for _, it := range n.Items {
		ctx.KeywordLocation = it.Location
		before := len(ctx.Units())
		ok, err := it.Keyword.Evaluate(ctx, it.Data, instance)
		if err != nil {
			return false, err
		}
		if !ok {
			valid = false
			if len(ctx.Units()) == before { // keyword gave no unit of its own
				if err := ctx.Report("rule " + it.Keyword.Name + " failed"); err != nil {
					return false, err
				}
			}
		}
	}
	return valid, nil
}
