package validation_test

import (
	"testing"

	"github.com/neoface/jsonschema"
)

func TestEnumSchemaErrors(t *testing.T) {
	for _, s := range []string{`{"enum":1}`, `{"enum":"a"}`, `{"enum":{}}`} {
		if _, err := jsonschema.Compile([]byte(s)); err == nil {
			t.Errorf("%s: want schema error", s)
		}
	}
}

func TestEnumDeepEquality(t *testing.T) {
	s, err := jsonschema.Compile([]byte(`{"enum":[1,null,{"a":1,"b":[1,2]},[[1]],"x"]}`))
	if err != nil {
		t.Fatal(err)
	}
	for doc, want := range map[string]bool{
		`1.0`: true, `null`: true, `{"b":[1.0,2],"a":1}`: true, `[[1.0]]`: true, `"x"`: true,
		`2`: false, `false`: false, `{"a":1}`: false, `[1]`: false, `"1"`: false,
	} {
		if r, err := s.Validate([]byte(doc)); err != nil || r.Valid() != want {
			t.Errorf("%s: got %v %v", doc, r.Valid(), err)
		}
	}
	e, _ := jsonschema.Compile([]byte(`{"enum":[]}`))
	if r, _ := e.Validate([]byte(`1`)); r.Valid() {
		t.Error("empty enum matches nothing")
	}
}
