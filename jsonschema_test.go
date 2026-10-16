package jsonschema

import (
	"errors"
	"testing"

	"github.com/neoface/jsonschema/internal/jsonvalue"
)

var docs = []string{`null`, `true`, `1`, `"s"`, `[1,2]`, `{"a":{}}`}

func TestValidateSchemas(t *testing.T) {
	cases := map[string]bool{`true`: true, `false`: false, `{}`: true, `{"type":"string"}`: true, `{"foo":1}`: true}
	for sch, want := range cases {
		s, err := Compile([]byte(sch))
		if err != nil {
			t.Fatal(sch, err)
		}
		for _, d := range docs {
			r, err := s.Validate([]byte(d))
			if err != nil || r.Valid() != want {
				t.Errorf("%s vs %s: %v %v", sch, d, r.Valid(), err)
			}
		}
	}
}

func TestCompileErrors(t *testing.T) {
	for _, sch := range []string{`5`, `"x"`, `null`, `[]`, ``, `{`, `{} {}`, `{"a":1,"a":2}`} {
		if _, err := Compile([]byte(sch)); err == nil {
			t.Errorf("%q: want error", sch)
		}
	}
}

func TestValidateBrokenDocument(t *testing.T) {
	s, _ := Compile([]byte(`true`))
	for _, d := range []string{``, `{`, `{} {}`, `{"a":1,"x":{"a":1,"a":2}}`} {
		if _, err := s.Validate([]byte(d)); err == nil {
			t.Errorf("%q: want error", d)
		}
	}
	var se *jsonvalue.SyntaxError
	_, err := s.Validate([]byte(`{`))
	if !errors.As(err, &se) {
		t.Errorf("want SyntaxError, got %v", err)
	}
}
