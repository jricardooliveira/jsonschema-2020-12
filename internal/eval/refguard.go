package eval

import (
	"fmt"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/keywords"
)

// LoopError is returned when a schema is re-entered at the same instance
// location on the current evaluation path (core 8.2.3.1 recursion that never
// descends into the document). It names locations only, never instance data.
type LoopError struct {
	SchemaLocation   string
	InstanceLocation string
}

func (e *LoopError) Error() string {
	return fmt.Sprintf("reference loop: schema location #%s re-entered at instance location %q", e.SchemaLocation, e.InstanceLocation)
}

type frame struct {
	node *compile.Node
	inst string
}

// callState is the per-call Items record plus the loop guard's active path.
// It embeds the record so every keyword's interface assertions on ctx.Items
// still hold. It lives in one evaluation, never in a compiled schema.
type callState struct {
	*EvaluatedItems
	active map[frame]struct{}
}

// enterNode marks (n, instance location) active and returns the function that
// removes it when the frame exits. Re-entry is a *LoopError.
func enterNode(ctx *keywords.EvalContext, n *compile.Node) (func(), error) {
	st, ok := ctx.Items.(*callState)
	if !ok {
		rec, isRec := ctx.Items.(*EvaluatedItems)
		if !isRec {
			return func() {}, nil
		}
		// First frame of a call: wrap the record (same pointer, so callers
		// that kept it still see every annotation).
		st = &callState{rec, map[frame]struct{}{}}
		ctx.Items = st
	}
	f := frame{n, ctx.InstanceLocation}
	if _, dup := st.active[f]; dup {
		return nil, &LoopError{SchemaLocation: n.Location, InstanceLocation: ctx.InstanceLocation}
	}
	st.active[f] = struct{}{}
	return func() { delete(st.active, f) }, nil
}
