package keywords

import "testing"

func newParent() *EvalContext {
	return NewEvalContext(100, func(*EvalContext, any, any) (bool, error) { return true, nil })
}

func TestChildStackReusesPerDepth(t *testing.T) {
	var s childStack
	a, b := s.at(0), s.at(0)
	if a != b {
		t.Fatal("siblings at one depth must share a context")
	}
	if s.at(1) == a {
		t.Fatal("different depths must not share a context")
	}
	if len(s.ctxs) != 2 {
		t.Fatalf("stack grew to %d, want 2", len(s.ctxs))
	}
}

func TestChildResetClearsPerChildState(t *testing.T) {
	p := newParent()
	p.Items = "items"
	var s childStack
	c := s.at(0)
	c.reset(p, "/a", &TreeNode{})
	c.Keyword, c.Scope, c.Loc = "kw", 7, nil
	c.units = append(c.units, Unit{Error: "x"})

	c.reset(p, "/b", nil)
	if c.InstanceLocation != "/b" || c.Tree != nil || c.Keyword != "" || c.Scope != 0 ||
		c.Loc != nil || len(c.units) != 0 {
		t.Fatalf("state leaked after reset: %+v", c)
	}
	if c.Items != "items" || c.count != p.count || c.col != p.col || c.max != p.max || c.evalNode == nil {
		t.Fatal("shared per-call state must be inherited from the parent")
	}
}

func TestChildResetDoesNotAliasReturnedUnits(t *testing.T) {
	p := newParent()
	var s childStack
	c := s.at(0)
	c.reset(p, "/a", nil)
	c.units = append(c.units, Unit{Error: "first"})
	returned := c.units // what EvalChild hands to the caller

	c.reset(p, "/b", nil)
	c.units = append(c.units, Unit{Error: "second"})
	if returned[0].Error != "first" {
		t.Fatalf("returned units overwritten: %q", returned[0].Error)
	}
}
