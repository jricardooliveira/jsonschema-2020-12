package eval

import (
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
)

func TestEvaluateNoAnnotationsRecordsNone(t *testing.T) {
	n, err := compile.Compile(map[string]any{"title": "t", "type": "string"})
	if err != nil {
		t.Fatal(err)
	}
	lim := DefaultLimits()
	with, err := EvaluateWith(n, "x", lim, false)
	if err != nil || !with.Valid || len(with.Annotations) == 0 {
		t.Fatalf("control: want annotations, got %+v %v", with, err)
	}
	ev, err := EvaluateNoAnnotations(n, "x", lim)
	if err != nil || !ev.Valid || len(ev.Annotations) != 0 {
		t.Fatalf("got %+v %v", ev, err)
	}
	bad, err := EvaluateNoAnnotations(n, 1, lim)
	if err != nil || bad.Valid || len(bad.Units) == 0 {
		t.Fatalf("invalid case: %+v %v", bad, err)
	}
}

func TestEvaluateNoAnnotationsStillCountsCap(t *testing.T) {
	n, err := compile.Compile(map[string]any{"title": "t"})
	if err != nil {
		t.Fatal(err)
	}
	lim := DefaultLimits()
	lim.MaxAnnotations = 1
	if _, err := EvaluateNoAnnotations(n, "x", lim); err != nil {
		t.Fatal(err)
	}
	n2, _ := compile.Compile(map[string]any{"title": "t", "description": "d"})
	if _, err := EvaluateNoAnnotations(n2, "x", lim); err == nil {
		t.Fatal("annotation cap not enforced")
	}
}
