package eval

import (
	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

// DefaultMaxOutputUnits caps how many units one evaluation may append,
// counting units that are later discarded.
const DefaultMaxOutputUnits = 10_000_000

// EvaluateUnits is Evaluate that also returns the failure units in
// deterministic depth-first order. Exceeding max appended units is a
// *jsonvalue.LimitError named "output units".
func EvaluateUnits(n *compile.Node, instance any, max int) (bool, []keywords.Unit, error) {
	ctx := keywords.NewEvalContext(max, evalNode)
	ok, err := evalNode(ctx, n, instance)
	if err != nil {
		return false, nil, err
	}
	return ok, ctx.Units(), nil
}
