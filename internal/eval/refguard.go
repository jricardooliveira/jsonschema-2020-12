package eval

import (
	"errors"
	"fmt"
	"strings"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
)

// MaxEvalDepth caps how many schema nodes may be nested on one evaluation
// path. It is the backstop for $ref recursion that descends into the
// document; past it, evaluation returns a *jsonvalue.LimitError named
// "evaluation depth" instead of overflowing the Go stack. Recursion over the
// deepest document the decoder accepts (jsonvalue.MaxDepth) needs about two
// nodes per level, so this leaves ample room.
// ponytail: fixed bound; story 193 makes it an option and CLI flag.
const MaxEvalDepth = 4 * jsonvalue.MaxDepth

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

// pathFrame is one schema node on the active evaluation path.
type pathFrame struct {
	static, dynamic string // node location in the schema / along the evaluation path
	kw              string // name of the keyword being evaluated in this node
	kwDynamic       string // its location along the evaluation path
}

// callState is the per-call Items record plus the guard's active path. It
// embeds the record so every keyword's interface assertions on ctx.Items
// still hold. It lives in one evaluation, never in a compiled schema.
type callState struct {
	*EvaluatedItems
	active map[frame]struct{}
	path   []*pathFrame
}

// enterNode marks (n, instance location) active, pushes n on the evaluation
// path and returns the frame plus the function that undoes both when the
// node exits. Each entry counts one evaluation step. Re-entry is a
// *LoopError; a path deeper than MaxEvalDepth or more steps than the call's
// cap is a *jsonvalue.LimitError.
func enterNode(ctx *keywords.EvalContext, n *compile.Node) (*pathFrame, func(), error) {
	st, ok := ctx.Items.(*callState)
	if !ok {
		rec, isRec := ctx.Items.(*EvaluatedItems)
		if !isRec { // fail closed: never evaluate without the guard
			return nil, nil, errors.New("internal error: evaluation state missing")
		}
		// First frame of a call: wrap the record (same pointer, so callers
		// that kept it still see every annotation).
		st = &callState{EvaluatedItems: rec, active: map[frame]struct{}{}}
		ctx.Items = st
	}
	// Every node entry, including through $ref, is one evaluation step, so
	// exponential branching ends with the "evaluation steps" LimitError.
	if err := st.Step(); err != nil {
		return nil, nil, err
	}
	if len(st.path) >= MaxEvalDepth {
		return nil, nil, &jsonvalue.LimitError{Limit: "evaluation depth", Max: MaxEvalDepth}
	}
	f := frame{n, ctx.InstanceLocation}
	if _, dup := st.active[f]; dup {
		return nil, nil, &LoopError{SchemaLocation: n.Location, InstanceLocation: ctx.InstanceLocation}
	}
	pf := &pathFrame{static: n.Location, dynamic: n.Location}
	if k := len(st.path); k > 0 {
		pf.dynamic = dynamicLocation(st.path[k-1], n.Location)
	}
	st.active[f] = struct{}{}
	st.path = append(st.path, pf)
	return pf, func() {
		delete(st.active, f)
		st.path = st.path[:len(st.path)-1]
	}, nil
}

// dynamicLocation is the evaluation-path location of a node with static
// location loc, entered from parent (core 12.3.1: keywordLocation follows
// $ref). A reference target starts at the $ref keyword itself; any other
// child keeps its path below the parent node.
func dynamicLocation(parent *pathFrame, loc string) string {
	if parent.kw == "$ref" {
		return parent.kwDynamic
	}
	if rel, ok := strings.CutPrefix(loc, parent.static); ok && (rel == "" || rel[0] == '/') {
		return parent.dynamic + rel
	}
	return parent.kwDynamic
}

// at sets the keyword being evaluated in this node and returns its location
// along the evaluation path.
func (pf *pathFrame) at(kw, static string) string {
	pf.kw = kw
	pf.kwDynamic = pf.dynamic + strings.TrimPrefix(static, pf.static)
	return pf.kwDynamic
}
