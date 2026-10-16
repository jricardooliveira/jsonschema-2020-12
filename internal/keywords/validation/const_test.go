package validation_test

import (
	"testing"

	"github.com/neoface/jsonschema"
)

func TestConstDeepEquality(t *testing.T) {
	cases := []struct {
		schema string
		doc    string
		want   bool
	}{
		{`{"const":1}`, `1.0`, true},
		{`{"const":1}`, `true`, false},
		{`{"const":null}`, `null`, true},
		{`{"const":null}`, `0`, false},
		{`{"const":{"a":1,"b":2}}`, `{"b":2,"a":1}`, true},
		{`{"const":{"a":1}}`, `{"a":1,"b":2}`, false},
		{`{"const":[[1,2]]}`, `[[1,2.0]]`, true},
		{`{"const":[1,2]}`, `[2,1]`, false},
		{`{"const":12345678901234567890}`, `12345678901234567891`, false},
	}
	for _, c := range cases {
		s, err := jsonschema.Compile([]byte(c.schema))
		if err != nil {
			t.Fatal(err)
		}
		if r, err := s.Validate([]byte(c.doc)); err != nil || r.Valid() != c.want {
			t.Errorf("%s vs %s: got %v %v", c.schema, c.doc, r.Valid(), err)
		}
	}
}
