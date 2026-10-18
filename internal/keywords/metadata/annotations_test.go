package metadata

import (
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

func TestKeywordsAnnotateAndNeverFail(t *testing.T) {
	for _, n := range []string{"title", "description", "default", "deprecated", "readOnly", "writeOnly", "examples"} {
		k, ok := keywords.Lookup(n)
		if !ok || k.Vocabulary != VocabularyURI {
			t.Fatalf("%s not registered in vocabulary", n)
		}
		for _, v := range []any{"x", nil, 3.5, []any{"a"}, map[string]any{"type": "string"}} {
			data, err := k.Compile(&keywords.CompileContext{}, v)
			if err != nil {
				t.Fatalf("%s: %v", n, err)
			}
			ctx := keywords.NewEvalContext(10, nil)
			ctx.Keyword = n
			valid, err := k.Evaluate(ctx, data, 42)
			if !valid || err != nil {
				t.Fatalf("%s: valid=%v err=%v", n, valid, err)
			}
			if a := ctx.Annotations(); len(a) != 1 || a[0].Keyword != n {
				t.Fatalf("%s: got %+v", n, a)
			}
		}
	}
}
