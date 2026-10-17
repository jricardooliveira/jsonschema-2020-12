package eval

import (
	"errors"
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/jsonvalue"
)

const sinkSchema = `{"not":{"type":"string","x-in":1,"title":"t"},"x-out":2}`

func sinkNode(t *testing.T, schema string) *compile.Node {
	t.Helper()
	s, err := jsonvalue.DecodeLimited([]byte(schema), jsonvalue.NumberLimits{MaxDigits: 1000, MaxExponent: 1000})
	if err != nil {
		t.Fatal(err)
	}
	n, err := compile.Compile(s)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestNotSubtreeRecordsNothing(t *testing.T) {
	n := sinkNode(t, sinkSchema)
	ev, err := EvaluateLazy(n, 5, DefaultLimits())
	if err != nil || !ev.Valid {
		t.Fatal(ev, err)
	}
	anns := ev.Pending.Sorted() // lazy: chunks are flattened on first read
	if len(anns) != 1 || anns[0].Keyword != "x-out" {
		t.Fatalf("got %+v", anns)
	}
	// the tree path never discards: the same child annotates there
	tr, err := EvaluateWith(n, 5, DefaultLimits(), true)
	if err != nil || len(tr.Annotations) != 1 {
		t.Fatal(tr, err)
	}
}

func TestDiscardedAnnotationsStillCountAgainstCap(t *testing.T) {
	n := sinkNode(t, sinkSchema)
	lim := DefaultLimits()
	lim.MaxAnnotations = 3 // x-in, title (dropped) and x-out
	if _, err := EvaluateLazy(n, 5, lim); err != nil {
		t.Fatal(err)
	}
	lim.MaxAnnotations = 2
	var le *jsonvalue.LimitError
	if _, err := EvaluateLazy(n, 5, lim); !errors.As(err, &le) || le.Limit != "annotations" {
		t.Fatalf("want annotations limit, got %v", err)
	}
}

func TestDiscardModeChargesSameSteps(t *testing.T) {
	n := sinkNode(t, sinkSchema)
	minSteps := func(tree bool) int {
		lim := DefaultLimits()
		for s := 1; s < 1000; s++ {
			lim.MaxSteps = s
			if _, err := EvaluateWith(n, 5, lim, tree); err == nil {
				return s
			}
		}
		t.Fatal("no step count passes")
		return 0
	}
	if a, b := minSteps(false), minSteps(true); a != b {
		t.Fatalf("steps differ: discard %d, tree %d", a, b)
	}
}
