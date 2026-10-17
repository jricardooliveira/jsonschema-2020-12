package eval

import (
	"reflect"
	"sync"
	"testing"

	"github.com/neoface/jsonschema/internal/jsonvalue"
)

const pendingSchema = `{"properties":{"a":{"title":"A"},"b":{"title":"B"},"c":{"title":"C"}},"title":"root"}`

func pendingDoc(t *testing.T) any {
	t.Helper()
	v, err := jsonvalue.DecodeLimited([]byte(`{"a":1,"b":2,"c":3}`), jsonvalue.NumberLimits{MaxDigits: 1000, MaxExponent: 1000})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// A Validate evaluation hands over the chunks unflattened: nothing is
// flattened or sorted until the first Sorted call.
func TestEvaluateValidateDefersFlatten(t *testing.T) {
	ev, err := EvaluateValidate(sinkNode(t, pendingSchema), pendingDoc(t), DefaultLimits())
	if err != nil || !ev.Valid {
		t.Fatal(ev, err)
	}
	if ev.Annotations != nil {
		t.Fatal("annotations were flattened at Validate")
	}
	p := ev.Pending
	if p == nil || len(p.chunks) == 0 || p.sorted != nil {
		t.Fatalf("want pending chunks, got %+v", p)
	}
}

// Concurrent first readers: one flatten and sort inside the Once, the chunks
// released in it, every reader the same answer as the eager sorted path.
func TestPendingSortedConcurrentReleasesChunks(t *testing.T) {
	n := sinkNode(t, pendingSchema)
	full, err := EvaluateFull(n, pendingDoc(t), DefaultMaxOutputUnits, DefaultMaxAnnotations)
	if err != nil {
		t.Fatal(err)
	}
	ev, err := EvaluateValidate(n, pendingDoc(t), DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	const k = 16
	got := make([][]Annotation, k)
	var wg sync.WaitGroup
	for i := range k {
		wg.Add(1)
		go func() { defer wg.Done(); got[i] = ev.Pending.Sorted() }()
	}
	wg.Wait()
	for i := range k {
		if !reflect.DeepEqual(got[i], full.Annotations) {
			t.Fatalf("reader %d: %+v, want %+v", i, got[i], full.Annotations)
		}
	}
	if ev.Pending.chunks != nil {
		t.Fatal("chunks not released after the first Sorted")
	}
}

func TestPendingNilAndFailing(t *testing.T) {
	var p *Pending
	if p.Sorted() != nil {
		t.Fatal("nil Pending must have no annotations")
	}
	ev, err := EvaluateValidate(sinkNode(t, `{"type":"string","title":"t"}`), 5.0, DefaultLimits())
	if err != nil || ev.Valid {
		t.Fatal(ev, err)
	}
	if ev.Pending.Sorted() != nil {
		t.Fatal("a failing result keeps no annotations")
	}
}
