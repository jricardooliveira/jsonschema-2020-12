package formatannotation

import (
	"testing"

	"github.com/neoface/jsonschema/internal/keywords"
)

func TestFormatAnnotatesAndNeverFails(t *testing.T) {
	k, ok := keywords.Lookup("format")
	if !ok || k.Vocabulary != VocabularyURI {
		t.Fatal("format not registered in vocabulary")
	}
	for _, v := range []any{"email", "no-such-format", nil, 3.5, []any{"a"}} {
		data, err := k.Compile(&keywords.CompileContext{}, v)
		if err != nil {
			t.Fatal(err)
		}
		ctx := keywords.NewEvalContext(10, nil)
		ctx.Keyword = "format"
		valid, err := k.Evaluate(ctx, data, "not an email")
		if !valid || err != nil {
			t.Fatalf("valid=%v err=%v", valid, err)
		}
		if a := ctx.Annotations(); len(a) != 1 || a[0].Keyword != "format" {
			t.Fatalf("got %+v", a)
		}
	}
}
