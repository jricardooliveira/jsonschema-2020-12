package eval

import (
	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/keywords"
)

// DynamicScope is the stack of schema nodes entered on the current
// evaluation path, outermost first (core 8.2.3.2). It is per-call state;
// entries are read-only pointers into the compiled schema.
// ponytail: every entered node is pushed, not only resource roots; callers
// filter with the match predicate. Linear in depth, bounded by MaxEvalDepth.
type DynamicScope struct {
	entries []*compile.Node
}

// push adds n and returns the function that removes it again.
func (s *DynamicScope) push(n *compile.Node) func() {
	s.entries = append(s.entries, n)
	k := len(s.entries) - 1
	return func() { s.entries = s.entries[:k] }
}

// Len is the number of entered nodes.
func (s *DynamicScope) Len() int { return len(s.entries) }

// Outermost returns the first (outermost) entered node for which match is
// true, or nil.
func (s *DynamicScope) Outermost(match func(*compile.Node) bool) *compile.Node {
	for _, n := range s.entries {
		if match(n) {
			return n
		}
	}
	return nil
}

// ScopeOf returns the dynamic scope of the call ctx belongs to, or nil
// before the first node was entered.
func ScopeOf(ctx *keywords.EvalContext) *DynamicScope {
	if st, ok := ctx.Items.(*callState); ok {
		return &st.scope
	}
	return nil
}
