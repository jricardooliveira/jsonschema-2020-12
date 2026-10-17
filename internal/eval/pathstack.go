package eval

import (
	"errors"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
)

// Path frame stack.
//
// Invariant: a *pathFrame (and the lazyLoc embedded in it) is valid only while
// its node is being evaluated, that is between the callState.enter that
// returned it and the matching callState.exit. Frames are entered and left
// strictly last-in first-out, always through those explicit calls (never a
// closure), so the frame of depth d is reused by the next node entered at
// depth d. Nothing may keep a frame or its lazyLoc after exit; strings are
// rendered when a unit or annotation is created. The stack belongs to one
// callState, which lives in one evaluation: it is never pooled or shared
// between calls. Frame pointers are stable (the slice holds pointers), so a
// parent's frame stays valid while deeper frames are entered.

// callState returns the per-call state installed in ctx, installing it around
// the bare record on first use. A missing record fails closed.
func stateOf(ctx *keywords.EvalContext) (*callState, error) {
	if st, ok := ctx.Items.(*callState); ok {
		return st, nil
	}
	rec, isRec := ctx.Items.(*EvaluatedItems)
	if !isRec { // fail closed: never evaluate without the guard
		return nil, errors.New("internal error: evaluation state missing")
	}
	// First frame of a call: wrap the record (same pointer, so callers
	// that kept it still see every annotation).
	st := &callState{EvaluatedItems: rec, active: map[frame]struct{}{}, maxRefChain: DefaultMaxRefChain, maxPath: MaxEvalDepth}
	ctx.Items = st
	return st, nil
}

// enter marks (n, instance location) active and pushes n on the path. Each
// entry counts one evaluation step. Re-entry is a *LoopError; a path deeper
// than maxPath or more steps than the call's cap is a *jsonvalue.LimitError.
// On success the caller must call exit with the returned frame.
func (st *callState) enter(ctx *keywords.EvalContext, n *compile.Node) (*pathFrame, error) {
	// Every node entry, including through $ref, is one evaluation step, so
	// exponential branching ends with the "evaluation steps" LimitError.
	if err := st.Step(); err != nil {
		return nil, err
	}
	d := st.depth
	if d >= st.maxPath {
		return nil, &jsonvalue.LimitError{Limit: "nesting depth", Max: st.maxPath / 4}
	}
	f := frame{n, ctx.InstanceLocation}
	if _, dup := st.active[f]; dup {
		return nil, &LoopError{SchemaLocation: n.Location, InstanceLocation: ctx.InstanceLocation, refMax: st.maxRefChain}
	}
	dynamic, viaRef, jump := n.Location, false, false
	if d > 0 {
		parent := st.frames[d-1]
		dynamic = dynamicLocation(parent, n.Location)
		jump = parent.kw == "$ref" || parent.kw == "$dynamicRef"
		viaRef = parent.viaRef || jump
	}
	if jump {
		if st.refs >= st.maxRefChain {
			return nil, &jsonvalue.LimitError{Limit: "reference chain", Max: st.maxRefChain}
		}
		st.refs++
	}
	if d == len(st.frames) {
		st.frames = append(st.frames, new(pathFrame))
	}
	pf := st.frames[d]
	*pf = pathFrame{static: n.Location, dynamic: dynamic, viaRef: viaRef, key: f, jump: jump}
	pf.loc = lazyLoc{pf: pf, res: n.Resource}
	st.active[f] = struct{}{}
	st.scope.entries = append(st.scope.entries, n)
	st.depth++
	return pf, nil
}

// exit undoes the matching enter. pf must be the frame of the innermost
// entered node.
func (st *callState) exit(pf *pathFrame) {
	st.scope.entries = st.scope.entries[:len(st.scope.entries)-1]
	if pf.jump {
		st.refs--
	}
	delete(st.active, pf.key)
	st.depth--
}
