package jsonschema

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/neoface/jsonschema/internal/jsonvalue"
)

// A large default under items, applied to many items, overruns a lowered
// report cap; the same input renders under DefaultMaxReportBytes.
func TestReportBytesCap(t *testing.T) {
	schema := []byte(`{"items":{"default":"` + strings.Repeat("d", 1000) + `"}}`)
	doc := []byte("[" + strings.TrimSuffix(strings.Repeat("0,", 200), ",") + "]")
	s, err := CompileWithOptions(schema, Options{MaxReportBytes: 50_000})
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Validate(doc)
	if err != nil || !r.Valid() {
		t.Fatal(r, err)
	}
	b, err := r.Output("basic")
	var le *jsonvalue.LimitError
	if !errors.As(err, &le) || le.Limit != "report bytes" || le.Max != 50_000 || b != nil {
		t.Fatalf("want report bytes LimitError and nil bytes, got %d bytes, %v", len(b), err)
	}
	s, _ = Compile(schema)
	r, _ = s.Validate(doc)
	if b, err := r.Output("basic"); err != nil || len(b) <= 50_000 || len(b) > DefaultMaxReportBytes {
		t.Fatalf("default cap: %d bytes, %v", len(b), err)
	}
	if _, err := CompileWithOptions(schema, Options{MaxReportBytes: -1}); err == nil || err.Error() != "options: MaxReportBytes must not be negative" {
		t.Fatalf("negative MaxReportBytes: %v", err)
	}
	// A zero-value Result falls back to the default cap, never unlimited.
	if b, err := (Result{}).Output("flag"); err != nil || string(b) != `{"valid":false}` {
		t.Fatalf("zero Result: %s %v", b, err)
	}
}

// One Schema used from many goroutines gives the same annotations and
// report each time (run with -race).
func TestConcurrentValidateAnnotations(t *testing.T) {
	s, err := Compile([]byte(`{"title":"T","properties":{"a":{"default":7},"b":{"anyOf":[{"type":"string","x-a":1},{"x-b":[2]}]}},"items":{"examples":[1.5]}}`))
	if err != nil {
		t.Fatal(err)
	}
	docs := [][]byte{[]byte(`{"a":1,"b":2}`), []byte(`[1,2,3]`), []byte(`{"b":"s"}`)}
	type out struct {
		ann []Annotation
		rep []byte
	}
	want := make([]out, len(docs))
	for i, d := range docs {
		r, err := s.Validate(d)
		if err != nil {
			t.Fatal(err)
		}
		rep, err := r.Output("basic")
		if err != nil {
			t.Fatal(err)
		}
		want[i] = out{r.Annotations(), rep}
	}
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < 50; n++ {
				i := n % len(docs)
				r, err := s.Validate(docs[i])
				if err != nil {
					t.Error(err)
					return
				}
				rep, err := r.Output("basic")
				if err != nil || !bytes.Equal(rep, want[i].rep) || !reflect.DeepEqual(r.Annotations(), want[i].ann) {
					t.Errorf("doc %d differs: %s %v", i, rep, err)
					return
				}
			}
		}()
	}
	wg.Wait()
}
