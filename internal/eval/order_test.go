package eval

import (
	"slices"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

var orderLog []string

func init() {
	rec := func(name string, after bool) {
		keywords.Register(keywords.Keyword{
			Name: name, Vocabulary: "urn:test", AfterSiblings: after,
			Compile: func(*keywords.CompileContext, any) (any, error) { return nil, nil },
			Evaluate: func(*keywords.EvalContext, any, any) (bool, error) {
				orderLog = append(orderLog, name)
				return true, nil
			},
		})
	}
	rec("a-order-late", true) // sorts first, must still run last
	rec("m-order-plain", false)
	rec("z-order-late", true) // sorts last
	rec("z-order-plain", false)
}

func TestAfterSiblingsRunLastWhateverKeyOrder(t *testing.T) {
	n := mustCompile(t, map[string]any{
		"z-order-late": 1, "a-order-late": 1, "z-order-plain": 1, "m-order-plain": 1,
	})
	orderLog = nil
	if ok, err := Evaluate(n, "x"); !ok || err != nil {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if len(orderLog) != 4 {
		t.Fatalf("log=%v", orderLog)
	}
	for _, late := range orderLog[:2] {
		if late == "a-order-late" || late == "z-order-late" {
			t.Fatalf("flagged keyword ran before a sibling: %v", orderLog)
		}
	}
	if !slices.Contains(orderLog[2:], "a-order-late") || !slices.Contains(orderLog[2:], "z-order-late") {
		t.Fatalf("log=%v", orderLog)
	}
}
