package eval

import (
	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
)

// DefaultMaxEvalSteps caps the evaluation steps of one call.
// ponytail: only the object keywords count steps so far; story 193 makes it
// an option and CLI flag and has every keyword count.
const DefaultMaxEvalSteps = 100_000_000

// Step counts one evaluation step (a member visited, a pattern match attempt
// or a reported failure). Past the cap it returns a *jsonvalue.LimitError
// named "evaluation steps".
func (r *EvaluatedItems) Step() error {
	r.steps++
	if r.steps > r.maxSteps {
		return &jsonvalue.LimitError{Limit: "evaluation steps", Max: r.maxSteps}
	}
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
