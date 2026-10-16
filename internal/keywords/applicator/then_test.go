package applicator_test

import (
	"testing"

	"github.com/neoface/jsonschema/internal/eval"
)

func TestThenAppliesOnlyWhenIfPasses(t *testing.T) {
	n := mustCompile(t, `{"if":{"type":"integer"},"then":{"minimum":10}}`)
	for doc, want := range map[string]bool{`5`: false, `50`: true, `"a"`: true} {
		ok, err := eval.Evaluate(n, mustDecode(t, doc))
		if err != nil || ok != want {
			t.Errorf("%s: ok=%v err=%v", doc, ok, err)
		}
	}
}
