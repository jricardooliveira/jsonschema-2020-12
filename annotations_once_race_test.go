package jsonschema

import (
	"reflect"
	"sync"
	"testing"
)

// TestAnnotationsOnceConcurrent calls Annotations and Output concurrently on
// one shared Result: the lazy flatten and sort runs once, every caller sees
// the same answer, and -race finds no data race.
func TestAnnotationsOnceConcurrent(t *testing.T) {
	s, err := Compile([]byte(`{"properties":{"a":{"title":"A"},"b":{"title":"B"},"c":{"title":"C"}},"title":"root"}`))
	if err != nil {
		t.Fatal(err)
	}
	res, err := s.Validate([]byte(`{"a":1,"b":2,"c":3}`))
	if err != nil || !res.Valid() {
		t.Fatalf("valid=%v err=%v", res.Valid(), err)
	}
	const n = 16
	anns := make([][]Annotation, n)
	outs := make([]any, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			anns[i] = res.Annotations()
			o, err := res.Output("basic")
			if err != nil {
				t.Error(err)
			}
			outs[i] = o
		}()
	}
	wg.Wait()
	if len(anns[0]) == 0 {
		t.Fatal("expected annotations")
	}
	for i := 1; i < n; i++ {
		if !reflect.DeepEqual(anns[0], anns[i]) {
			t.Fatalf("Annotations differ between callers 0 and %d", i)
		}
		if !reflect.DeepEqual(outs[0], outs[i]) {
			t.Fatalf("Output differs between callers 0 and %d", i)
		}
	}
}
