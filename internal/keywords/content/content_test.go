package content

import (
	"testing"

	"github.com/neoface/jsonschema/internal/keywords"
)

func TestContentKeywordsAnnotateAndNeverFail(t *testing.T) {
	for _, n := range []string{"contentEncoding", "contentMediaType", "contentSchema"} {
		k, ok := keywords.Lookup(n)
		if !ok || k.Vocabulary != VocabularyURI {
			t.Fatalf("%s not registered in vocabulary", n)
		}
		// contentSchema {"type":"integer"} must not be applied to the string instance.
		for _, v := range []any{"base64", nil, 3.5, map[string]any{"type": "integer"}} {
			data, err := k.Compile(&keywords.CompileContext{}, v)
			if err != nil {
				t.Fatalf("%s: %v", n, err)
			}
			ctx := keywords.NewEvalContext(10, nil)
			ctx.Keyword = n
			valid, err := k.Evaluate(ctx, data, "text")
			if !valid || err != nil {
				t.Fatalf("%s: valid=%v err=%v", n, valid, err)
			}
			if a := ctx.Annotations(); len(a) != 1 || a[0].Keyword != n || len(ctx.Units()) != 0 {
				t.Fatalf("%s: got %+v", n, a)
			}
		}
	}
}
