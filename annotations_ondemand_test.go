package jsonschema_test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/neoface/jsonschema"
)

const onDemandSchema = `{
  "$defs": {"n": {"title": "num", "type": "integer"}},
  "title": "root",
  "properties": {"a": {"$ref": "#/$defs/n"}, "b/c": {"description": "d"}, "a~": {"default": 1}},
  "prefixItems": [{"title": "first"}],
  "not": {"type": "string", "title": "dropped"},
  "x-unknown": 7
}`

const onDemandDoc = `{"a": 1, "b/c": "s", "a~": 2}`

var onDemandShapes = []string{"flag", "basic", "detailed", "verbose"}

func TestAnnotationsOnDemandContentAndOrder(t *testing.T) {
	s, err := jsonschema.Compile([]byte(onDemandSchema))
	if err != nil {
		t.Fatal(err)
	}
	res, err := s.Validate([]byte(onDemandDoc))
	if err != nil || !res.Valid() {
		t.Fatal(res.Valid(), err)
	}
	anns := res.Annotations()
	if len(anns) == 0 {
		t.Fatal("no annotations")
	}
	for _, a := range anns {
		if a.Value == "dropped" || a.Keyword == "" {
			t.Fatalf("unexpected annotation %+v", a)
		}
	}
	// the root annotations come first (empty instance location sorts first)
	if anns[0].InstanceLocation != "" {
		t.Fatalf("first annotation not at root: %+v", anns[0])
	}
	// repeated access returns the same content
	if !reflect.DeepEqual(anns, res.Annotations()) {
		t.Fatal("second access differs")
	}
}

func TestResultConcurrentReadsAreRaceFree(t *testing.T) {
	s, err := jsonschema.Compile([]byte(onDemandSchema))
	if err != nil {
		t.Fatal(err)
	}
	res, err := s.Validate([]byte(onDemandDoc))
	if err != nil {
		t.Fatal(err)
	}
	wantAnn := func() []jsonschema.Annotation {
		r, _ := s.Validate([]byte(onDemandDoc))
		return r.Annotations()
	}()
	want := map[string]string{}
	for _, sh := range onDemandShapes {
		b, err := res.Output(sh)
		if err != nil {
			t.Fatal(err)
		}
		want[sh] = string(b)
	}

	// fresh result, so the first access happens concurrently
	shared, err := s.Validate([]byte(onDemandDoc))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan string, 64)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !reflect.DeepEqual(shared.Annotations(), wantAnn) {
				errs <- "annotations differ"
			}
			for _, sh := range onDemandShapes {
				b, err := shared.Output(sh)
				if err != nil || string(b) != want[sh] {
					errs <- "output differs: " + sh
				}
			}
			// concurrent Validate on the shared Schema
			r, err := s.Validate([]byte(onDemandDoc))
			if err != nil || !reflect.DeepEqual(r.Annotations(), wantAnn) {
				errs <- "validate differs"
			}
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}
