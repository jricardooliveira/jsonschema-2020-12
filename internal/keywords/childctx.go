package keywords

// childStack keeps one reusable child EvalContext per evaluation depth. It is
// per-call state: it is created with the call's root context and never shared
// between Validate calls, so a compiled Schema stays immutable.
//
// Invariants (EvalChild relies on both):
//
//  1. No keyword may keep its child context after EvalChild returns. The
//     context at a depth is reset and reused for the next sibling, so a kept
//     pointer would silently see another child's state.
//  2. The caller owns the units slice EvalChild returns. reset never reuses a
//     context's units buffer (it drops the reference instead of truncating),
//     so units handed out for one child are never overwritten by a later
//     sibling at the same depth.
//
// One context per depth is enough because evaluation is strictly last-in
// first-out: a child at depth d is finished before the next child at depth d
// starts, and a grandchild uses depth d+1.
type childStack struct {
	ctxs []*EvalContext
}

// at returns the context for depth (0 is the first child level), growing the
// stack only when depth is a new maximum.
func (s *childStack) at(depth int) *EvalContext {
	for len(s.ctxs) <= depth {
		s.ctxs = append(s.ctxs, new(EvalContext))
	}
	return s.ctxs[depth]
}

// reset makes c a fresh child of parent for the instance at instLoc, with
// tree as its evaluation-tree node (nil when none). It overwrites the whole
// struct, so no per-child state (units, location, keyword, scope) can leak
// from a previous sibling.
func (c *EvalContext) reset(parent *EvalContext, instLoc string, tree *TreeNode) {
	*c = EvalContext{
		InstanceLocation: instLoc,
		Tree:             tree,
		Items:            parent.Items,
		count:            parent.count,
		col:              parent.col,
		max:              parent.max,
		evalNode:         parent.evalNode,
	}
}
