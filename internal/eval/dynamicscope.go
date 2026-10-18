package eval

import (
	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

// DynamicScope is the stack of schema nodes entered on the current
// evaluation path, outermost first (core 8.2.3.2). It is per-call state;
// entries are read-only pointers into the compiled schema.
// ponytail: every entered node is pushed, not only resource roots; callers
// filter with the match predicate. Linear in depth, bounded by MaxEvalDepth.
type DynamicScope struct {
	entries []*compile.Node
}

// enter adds n. Entries are left strictly last-in first-out by exit; the
// evaluation path uses this pair, with no closure.
func (s *DynamicScope) enter(n *compile.Node) {
	s.entries = append(s.entries, n)
}

// exit removes the innermost entry added by enter.
func (s *DynamicScope) exit() {
	s.entries = s.entries[:len(s.entries)-1]
}

// push is the compatibility entry point: it adds n and returns the function
// that removes it again. Evaluation itself uses enter and exit.
func (s *DynamicScope) push(n *compile.Node) func() {
	s.enter(n)
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

func hasDynamicAnchor(n *compile.Node, name string) bool {
	if n == nil {
		return false
	}
	for _, it := range n.Items {
		if it.Keyword.Name == "$dynamicAnchor" && it.Data == name {
			return true
		}
	}
	return false
}

// ResolveDynamic implements core 8.2.3.2: bookending when initial has no
// matching $dynamicAnchor; otherwise the matching anchor of the outermost
// resource in the dynamic scope, looked up in its compiled anchor table (the
// anchor need not have been evaluated). A resource is in scope once any of
// its nodes was entered. Linear in scope depth, no allocation.
func (s *callState) ResolveDynamic(initial any, name string) any {
	n, _ := initial.(*compile.Node)
	if !hasDynamicAnchor(n, name) {
		return initial
	}
	if m := s.scope.Outermost(func(c *compile.Node) bool { return c.Resource != nil && c.Resource.Dynamic[name] != nil }); m != nil {
		return m.Resource.Dynamic[name]
	}
	return initial
}

// ScopeOf returns the dynamic scope of the call ctx belongs to, or nil
// before the first node was entered.
func ScopeOf(ctx *keywords.EvalContext) *DynamicScope {
	if st, ok := ctx.Items.(*callState); ok {
		return &st.scope
	}
	return nil
}
