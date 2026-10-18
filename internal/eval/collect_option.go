package eval

import (
	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

// EvaluateNoAnnotations is EvaluateWith (no tree) for a caller that only needs
// the pass or fail answer and the error units, as the meta-schema check does.
// Annotations are counted against the cap, so the annotation limit triggers at
// the same inputs, but are neither stored nor sorted. Evaluated-location
// tracking, step charging and every other limit run unchanged.
func EvaluateNoAnnotations(n *compile.Node, instance any, lim Limits) (Evaluation, error) {
	ctx := keywords.NewEvalContext(lim.MaxUnits, evalNode)
	ctx.Items = &callState{
		EvaluatedItems: newEvaluatedItems(lim.MaxSteps), active: map[frame]struct{}{},
		maxRefChain: lim.MaxRefChain, maxPath: 4 * lim.MaxDepth,
	}
	ctx.SetMaxAnnotations(lim.MaxAnnotations)
	ctx.SetDiscard(true)
	ok, err := evalNode(ctx, n, instance)
	if err != nil {
		return Evaluation{}, err
	}
	return Evaluation{Valid: ok, Units: ctx.Units()}, nil
}
