package applicator_test

import (
	"testing"

	"github.com/neoface/jsonschema/internal/eval"
)

func TestElseAppliesOnlyWhenIfFails(t *testing.T) {
	n := mustCompile(t, `{"if":{"type":"integer"},"else":{"type":"string"}}`)
	for doc, want := range map[string]bool{`5`: true, `"a"`: true, `true`: false} {
		ok, err := eval.Evaluate(n, mustDecode(t, doc))
		if err != nil || ok != want {
			t.Errorf("%s: ok=%v err=%v", doc, ok, err)
		}
	}
}
