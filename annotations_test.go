package jsonschema

import (
	"errors"
	"testing"

	"github.com/neoface/jsonschema/internal/jsonvalue"
)

func TestAnnotationsPassingOrdered(t *testing.T) {
	s, err := Compile([]byte(`{"properties":{"b":{"title":"B"},"a":{"title":"A"}},"description":"n"}`))
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Validate([]byte(`{"a":1,"b":2}`))
	if err != nil || !r.Valid() {
		t.Fatal(r, err)
	}
	got := r.Annotations()
	if len(got) < 3 || got[0].InstanceLocation != "" {
		t.Fatalf("got %+v", got)
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].InstanceLocation > got[i].InstanceLocation {
			t.Fatalf("unordered: %+v", got)
		}
	}
	got[0].Keyword = "mutated"
	if r.Annotations()[0].Keyword == "mutated" {
		t.Fatal("accessor must return a copy")
	}
}

func TestAnnotationsEmptyOnFailure(t *testing.T) {
	s, _ := Compile([]byte(`{"title":"t","type":"string"}`))
	r, err := s.Validate([]byte(`1`))
	if err != nil || r.Valid() || len(r.Annotations()) != 0 {
		t.Fatal(r, err)
	}
}

func TestAnnotationLimit(t *testing.T) {
	schema := []byte(`{"title":"t","description":"d"}`)
	s, err := CompileWithOptions(schema, Options{MaxAnnotations: 1})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Validate([]byte(`1`))
	var le *jsonvalue.LimitError
	if !errors.As(err, &le) || le.Limit != "annotations" {
		t.Fatalf("want annotations LimitError, got %v", err)
	}
	if _, err := CompileWithOptions(schema, Options{MaxAnnotations: -1}); err == nil {
		t.Fatal("negative must be an error")
	}
	s, _ = CompileWithOptions(schema, Options{})
	if r, err := s.Validate([]byte(`1`)); err != nil || len(r.Annotations()) != 2 || DefaultMaxAnnotations <= 0 {
		t.Fatal(r, err)
	}
}
