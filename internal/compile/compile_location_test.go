package compile

import (
	"testing"

	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "x-loc-sub", Vocabulary: "test",
		Compile: func(cc *keywords.CompileContext, v any) (any, error) {
			return cc.Sub(v, "/0")
		},
		Evaluate: func(*keywords.EvalContext, any, any) (bool, error) { return true, nil }})
}

func TestSubCompilesAtRootedLocation(t *testing.T) {
	n, err := Compile(map[string]any{"x-loc-sub": map[string]any{"x-loc-sub": true}})
	if err != nil {
		t.Fatal(err)
	}
	outer := n.Items[0]
	if outer.Location != "/x-loc-sub" {
		t.Fatalf("outer %q", outer.Location)
	}
	child := outer.Data.(*Node)
	if child.Location != "/x-loc-sub/0" || child.Items[0].Location != "/x-loc-sub/0/x-loc-sub" {
		t.Fatalf("child %q %q", child.Location, child.Items[0].Location)
	}
	if g := child.Items[0].Data.(*Node); g.Location != "/x-loc-sub/0/x-loc-sub/0" {
		t.Fatalf("grandchild %q", g.Location)
	}
}
