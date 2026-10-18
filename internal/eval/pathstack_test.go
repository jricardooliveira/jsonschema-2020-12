package eval

import (
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

func newStackCtx() *keywords.EvalContext {
	ctx := keywords.NewEvalContext(DefaultMaxOutputUnits, evalNode)
	ctx.Items = newEvaluatedItems(DefaultLimits().MaxSteps)
	return ctx
}

// Frames are entered and left last-in first-out; a depth's frame is reused by
// the next node entered at that depth, and a parent's frame survives deeper
// entries.
func TestPathStackLIFOAndReuse(t *testing.T) {
	ctx := newStackCtx()
	n := mustCompile(t, true)
	st, err := stateOf(ctx)
	if err != nil {
		t.Fatal(err)
	}
	a, err := st.enter(ctx, n)
	if err != nil {
		t.Fatal(err)
	}
	ctx.InstanceLocation = "/x"
	b, err := st.enter(ctx, n)
	if err != nil {
		t.Fatal(err)
	}
	if a == b || st.depth != 2 || st.scope.Len() != 2 {
		t.Fatalf("depth %d, scope %d, same frame %v", st.depth, st.scope.Len(), a == b)
	}
	st.exit(b)
	c, err := st.enter(ctx, n)
	if err != nil {
		t.Fatal(err)
	}
	if c != b {
		t.Error("frame of depth 1 not reused")
	}
	st.exit(c)
	st.exit(a)
	if st.depth != 0 || st.scope.Len() != 0 || len(st.active) != 0 {
		t.Errorf("state not empty: depth %d scope %d active %d", st.depth, st.scope.Len(), len(st.active))
	}
}

// Re-entering the same node at the same instance location is a loop; the
// failed enter leaves the stack unchanged.
func TestPathStackLoopLeavesStackIntact(t *testing.T) {
	ctx := newStackCtx()
	n := mustCompile(t, true)
	st, _ := stateOf(ctx)
	if _, err := st.enter(ctx, n); err != nil {
		t.Fatal(err)
	}
	if _, err := st.enter(ctx, n); err == nil {
		t.Fatal("loop not detected")
	}
	if st.depth != 1 {
		t.Errorf("depth %d after failed enter", st.depth)
	}
}

// Each call's state owns its own stack.
func TestPathStackPerCallIsolation(t *testing.T) {
	n := mustCompile(t, true)
	c1, c2 := newStackCtx(), newStackCtx()
	s1, _ := stateOf(c1)
	s2, _ := stateOf(c2)
	f1, err := s1.enter(c1, n)
	if err != nil {
		t.Fatal(err)
	}
	f2, err := s2.enter(c2, n)
	if err != nil {
		t.Fatal(err)
	}
	if f1 == f2 || s1.depth != 1 || s2.depth != 1 {
		t.Error("calls share frames")
	}
	s1.exit(f1)
	if s2.depth != 1 {
		t.Error("exit in one call changed another")
	}
	s2.exit(f2)
}
