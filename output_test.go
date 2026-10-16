package jsonschema

import (
	"errors"
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/jsonvalue"
)

func TestOutputShapes(t *testing.T) {
	const secret = "s3cr3t-value"
	s, err := Compile([]byte(`{"type":"number"}`))
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Validate([]byte(`"` + secret + `"`))
	if err != nil {
		t.Fatal(err)
	}
	for shape, want := range map[string]string{
		"flag":  `{"valid":false}`,
		"basic": `{"valid":false,"errors":[{"instanceLocation":"","keywordLocation":"/type","error":"value must be of type number"}]}`,
	} {
		b, err := r.Output(shape)
		if err != nil || string(b) != want || strings.Contains(string(b), secret) {
			t.Errorf("%s: %s %v", shape, b, err)
		}
	}
	ok, _ := s.Validate([]byte(`1`))
	if b, _ := ok.Output("basic"); string(b) != `{"valid":true}` {
		t.Errorf("valid basic: %s", b)
	}
	if _, err := r.Output("fancy"); err == nil {
		t.Error("unknown shape must error")
	}
}

func TestOutputUnitCap(t *testing.T) {
	schema := []byte(`{"type":"string","const":"x"}`)
	s, err := CompileWithOptions(schema, Options{MaxOutputUnits: 1})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Validate([]byte(`1`))
	var le *jsonvalue.LimitError
	if !errors.As(err, &le) || le.Limit != "output units" {
		t.Fatalf("want output units LimitError, got %v", err)
	}
	s, _ = CompileWithOptions(schema, Options{MaxOutputUnits: DefaultMaxOutputUnits})
	if _, err := s.Validate([]byte(`1`)); err != nil {
		t.Fatal(err)
	}
	if _, err := CompileWithOptions(schema, Options{MaxOutputUnits: -1}); err == nil {
		t.Error("negative cap must error")
	}
}
