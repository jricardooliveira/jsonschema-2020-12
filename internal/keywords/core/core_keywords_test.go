package core_test

import (
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/eval"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	_ "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords/all"
)

func run(t *testing.T, schema, doc string) (bool, error) {
	t.Helper()
	s, err := jsonvalue.DecodeLimited([]byte(schema), jsonvalue.NumberLimits{})
	if err != nil {
		t.Fatal(err)
	}
	d, err := jsonvalue.DecodeLimited([]byte(doc), jsonvalue.NumberLimits{})
	if err != nil {
		t.Fatal(err)
	}
	n, err := compile.Compile(s)
	if err != nil {
		return false, err
	}
	return eval.Evaluate(n, d)
}

// core 8.3: $comment, and 8.2.4: $defs, never change a result.
func TestNoEffectKeywords(t *testing.T) {
	for _, s := range []string{
		`{"$comment":"x","type":"string"}`,
		`{"$defs":{"a":false},"type":"string"}`,
		`{"$id":"http://x/y","$anchor":"a","type":"string"}`,
	} {
		if ok, err := run(t, s, `"a"`); !ok || err != nil {
			t.Errorf("%s: %v %v", s, ok, err)
		}
		if ok, err := run(t, s, `1`); ok || err != nil {
			t.Errorf("%s on 1: %v %v", s, ok, err)
		}
	}
}

func TestWrongTypesAreSchemaErrors(t *testing.T) {
	for _, s := range []string{`{"$id":1}`, `{"$anchor":1}`, `{"$comment":{}}`, `{"$defs":[]}`, `{"$ref":1}`} {
		if _, err := run(t, s, `1`); err == nil {
			t.Errorf("%s accepted", s)
		}
	}
}
