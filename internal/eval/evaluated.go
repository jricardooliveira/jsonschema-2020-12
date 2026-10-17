package eval

import (
	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

// DefaultMaxEvalSteps caps the evaluation steps of one call.
// ponytail: only the object keywords count steps so far; story 193 makes it
// an option and CLI flag and has every keyword count.
const DefaultMaxEvalSteps = 100_000_000

// Step counts one evaluation step (a node entered, a keyword evaluated, a
// member visited or a reported failure).
func (r *EvaluatedItems) Step() error { return r.Charge(1) }

// Charge counts n evaluation steps. Past the cap it returns a
// *jsonvalue.LimitError named "evaluation steps". The comparison cannot
// overflow.
func (r *EvaluatedItems) Charge(n int) error {
	if n < 0 || r.steps > r.maxSteps-n {
		r.steps = r.maxSteps + 1
		return &jsonvalue.LimitError{Limit: "evaluation steps", Max: r.maxSteps}
	}
	r.steps += n
	return nil
}

// EvaluateLimited is EvaluateUnits with an explicit evaluation step cap.
func EvaluateLimited(n *compile.Node, instance any, maxUnits, maxSteps int) (bool, []keywords.Unit, error) {
	ctx := keywords.NewEvalContext(maxUnits, evalNode)
	ctx.Items = newEvaluatedItems(maxSteps)
	ok, err := evalNode(ctx, n, instance)
	if err != nil {
		return false, nil, err
	}
	return ok, ctx.Units(), nil
}

// RecordProperties records the member names a keyword applied a subschema to
// for the object at instLoc. kwLoc is the keyword location. The slice is
// copied. It shares the item log, so a failed branch's Rollback drops it too.
func (r *EvaluatedItems) RecordProperties(instLoc, kwLoc string, names []string) {
	if r.skip {
		return
	}
	r.entries = append(r.entries, itemEntry{
		key:   itemKey{instLoc, kwLoc},
		props: true,
		names: append([]string(nil), names...),
	})
}

// PropertiesAt returns the recorded names for the object at instLoc, keyed by
// keyword location. Slices are copies.
func (r *EvaluatedItems) PropertiesAt(instLoc string) map[string][]string {
	out := map[string][]string{}
	for _, e := range r.entries {
		if e.props && e.key.instance == instLoc {
			out[e.key.keyword] = append([]string(nil), e.names...)
		}
	}
	return out
}
