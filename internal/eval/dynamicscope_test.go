package eval

import (
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
)

func TestDynamicScopePushPop(t *testing.T) {
	var s DynamicScope
	a, b := &compile.Node{Location: "/a"}, &compile.Node{Location: "/b"}
	popA := s.push(a)
	popB := s.push(b)
	if s.Len() != 2 {
		t.Fatalf("len = %d", s.Len())
	}
	popB()
	if s.Len() != 1 || s.Outermost(func(*compile.Node) bool { return true }) != a {
		t.Fatal("pop must remove only the innermost entry")
	}
	popA()
	if s.Len() != 0 {
		t.Fatal("scope not empty after pops")
	}
}

func TestDynamicScopeOutermost(t *testing.T) {
	var s DynamicScope
	a, b := &compile.Node{Location: "/a"}, &compile.Node{Location: "/b"}
	s.push(a)
	s.push(b)
	if got := s.Outermost(func(*compile.Node) bool { return true }); got != a {
		t.Errorf("outermost = %v, want first entered", got)
	}
	if got := s.Outermost(func(n *compile.Node) bool { return n == b }); got != b {
		t.Errorf("match b = %v", got)
	}
	if got := s.Outermost(func(*compile.Node) bool { return false }); got != nil {
		t.Errorf("no match = %v, want nil", got)
	}
}

func TestDynamicScopePoppedAfterEvaluation(t *testing.T) {
	n, err := compile.Compile(map[string]any{"properties": map[string]any{"a": false}})
	if err != nil {
		t.Fatal(err)
	}
	// Invalid result still returns normally; the scope is per call, so a
	// second call starts from an empty scope (no leak between calls).
	for i := 0; i < 2; i++ {
		if _, err := Evaluate(n, map[string]any{"a": 1}); err != nil {
			t.Fatal(err)
		}
	}
}
