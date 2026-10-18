package eval

import (
	"fmt"
	"strings"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
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
	refMax           int // the call's reference-chain cap
}

// Unwrap makes a never-ending reference cycle match a reference-chain limit
// error (errors.As), carrying the configured cap.
func (e *LoopError) Unwrap() error {
	return &jsonvalue.LimitError{Limit: "reference chain", Max: e.refMax}
}

func (e *LoopError) Error() string {
	return fmt.Sprintf("reference loop: schema location #%s re-entered at instance location %q", e.SchemaLocation, e.InstanceLocation)
}

type frame struct {
	node *compile.Node
	inst string
}

// pathFrame is one schema node on the active evaluation path. See
// pathstack.go for its validity invariant.
type pathFrame struct {
	static, dynamic string // node location in the schema / along the evaluation path
	kw              string // name of the keyword being evaluated in this node
	kwDynamic       string // its location along the evaluation path
	viaRef          bool   // the path to this node crossed $ref or $dynamicRef
	jump            bool   // entering this node counted one reference jump
	key             frame  // the active-set key of this node
	loc             lazyLoc
}

// callState is the per-call Items record plus the guard's active path. It
// embeds the record so every keyword's interface assertions on ctx.Items
// still hold. It lives in one evaluation, never in a compiled schema.
type callState struct {
	*EvaluatedItems
	active map[frame]struct{}
	frames []*pathFrame // path frames indexed by depth, reused (pathstack.go)
	depth  int          // nodes currently entered
	scope  DynamicScope
	refs   int // $ref/$dynamicRef jumps on the current path
	// per-call caps: reference jumps and path depth (4 x the nesting depth)
	maxRefChain, maxPath int
}

// enterNode is the closure-returning entry point: it enters n and returns the
// frame plus the function that leaves it. Evaluation itself uses
// callState.enter and exit directly, without the closure.
func enterNode(ctx *keywords.EvalContext, n *compile.Node) (*pathFrame, func(), error) {
	st, err := stateOf(ctx)
	if err != nil {
		return nil, nil, err
	}
	pf, err := st.enter(ctx, n)
	if err != nil {
		return nil, nil, err
	}
	return pf, func() { st.exit(pf) }, nil
}

// dynamicLocation is the evaluation-path location of a node with static
// location loc, entered from parent (core 12.3.1: keywordLocation follows
// $ref). A reference target starts at the $ref keyword itself; any other
// child keeps its path below the parent node.
func dynamicLocation(parent *pathFrame, loc string) string {
	if parent.kw == "$ref" || parent.kw == "$dynamicRef" {
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
