package validation

import (
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

// TestMultipleOfSmallZeroAllocs: the yes/no decision on small numbers
// allocates nothing, for passing and failing instances (the error entry of a
// failing instance is not part of the decision).
func TestMultipleOfSmallZeroAllocs(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector adds allocations the validator did not cause")
	}
	div := compileForTest(t, mustNum(t, "0.01"))
	ctx := keywords.NewEvalContext(1000, nil)
	pass := mustNum(t, "12.99")
	fail := mustNum(t, "12.999")

	if ok, err := evalMultipleOf(ctx, div, pass); !ok || err != nil {
		t.Fatalf("12.99 multipleOf 0.01: ok=%v err=%v", ok, err)
	}
	if a := testing.AllocsPerRun(1000, func() { _, _ = evalMultipleOf(ctx, div, pass) }); a != 0 {
		t.Errorf("passing evalMultipleOf allocs = %v, want 0", a)
	}

	m, e, ok := fail.Inline()
	if !ok {
		t.Fatal("12.999 not small")
	}
	var multiple, decided bool
	if a := testing.AllocsPerRun(1000, func() { multiple, decided = multipleSmall(m, e, div.sm, div.se) }); a != 0 {
		t.Errorf("failing decision allocs = %v, want 0", a)
	}
	if multiple || !decided {
		t.Errorf("12.999 decision = (%v,%v), want (false,true)", multiple, decided)
	}
}
