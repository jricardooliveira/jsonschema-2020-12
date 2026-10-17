package eval

import (
	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/jsonvalue"
)

// DefaultMaxRefChain is how many $ref/$dynamicRef jumps one evaluation path
// may make: twice the default nesting depth.
const DefaultMaxRefChain = 2 * jsonvalue.MaxDepth

// Limits are the per-call caps of one evaluation. Every field must be
// positive (the caller resolves defaults). Counters built from them live only
// in the call's own state, so a compiled schema is never touched.
// A very large MaxDepth can exhaust the goroutine stack.
type Limits struct {
	MaxUnits, MaxAnnotations int
	// MaxSteps is the work-step budget; BytesPerStep in package keywords
	// scales the charge for strings and values.
	MaxSteps int
	// MaxRefChain caps $ref/$dynamicRef jumps on the current path.
	MaxRefChain int
	// MaxDepth is the nesting depth; the schema-node path backstop is 4 x it.
	MaxDepth int
}

// DefaultLimits returns the default caps.
func DefaultLimits() Limits {
	return Limits{DefaultMaxOutputUnits, DefaultMaxAnnotations, DefaultMaxEvalSteps, DefaultMaxRefChain, jsonvalue.MaxDepth}
}

// EvaluateWith evaluates n with a fresh budget built from lim. When tree is
// true it also builds the evaluation tree.
func EvaluateWith(n *compile.Node, instance any, lim Limits, tree bool) (Evaluation, error) {
	var t *Tree
	if tree {
		t = &Tree{}
	}
	return evaluateWith(n, instance, lim, t)
}

func evaluateWith(n *compile.Node, instance any, lim Limits, tree *Tree) (Evaluation, error) {
	items := newEvaluatedItems(lim.MaxSteps)
	return evaluate(n, instance, lim.MaxUnits, lim.MaxAnnotations, tree, &callState{
		EvaluatedItems: items, active: map[frame]struct{}{},
		maxRefChain: lim.MaxRefChain, maxPath: 4 * lim.MaxDepth,
	})
}
