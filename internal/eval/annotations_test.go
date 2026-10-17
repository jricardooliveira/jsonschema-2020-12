package eval

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	_ "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords/all"
)

func evalFull(t *testing.T, schema, doc string, maxAnn int) (Evaluation, error) {
	t.Helper()
	s, err := jsonvalue.DecodeLimited([]byte(schema), jsonvalue.NumberLimits{MaxDigits: 1000, MaxExponent: 1000})
	if err != nil {
		t.Fatal(err)
	}
	n, err := compile.Compile(s)
	if err != nil {
		t.Fatal(err)
	}
	d, err := jsonvalue.DecodeLimited([]byte(doc), jsonvalue.NumberLimits{MaxDigits: 1000, MaxExponent: 1000})
	if err != nil {
		t.Fatal(err)
	}
	return EvaluateTree(n, d, DefaultMaxOutputUnits, maxAnn)
}

// treeAnnotations collects every annotation held anywhere in the tree.
func treeAnnotations(tn *Tree) []Annotation {
	out := append([]Annotation(nil), tn.Annotations...)
	for _, c := range tn.Children {
		out = append(out, treeAnnotations(c)...)
	}
	return out
}

func TestPassingDescendantOfFailingNodeKeepsNoAnnotations(t *testing.T) {
	// anyOf/0 fails, but its allOf/0 grandchild passes and annotates x-g.
	ev, err := evalFull(t, `{"anyOf":[{"allOf":[{"x-g":1}],"type":"string"},{"x-b":2}]}`, `5`, DefaultMaxAnnotations)
	if err != nil || !ev.Valid {
		t.Fatal(ev, err)
	}
	got := treeAnnotations(ev.Tree)
	if len(got) != 1 || got[0].Keyword != "x-b" {
		t.Fatalf("dropped annotation still in tree: %+v", got)
	}
	branch := ev.Tree.Children[0]
	if branch.Valid || len(branch.Children) != 1 || !branch.Children[0].Valid {
		t.Fatalf("tree must keep the passing grandchild's result: %+v", branch)
	}
}

func TestEvaluateFullBuildsNoTree(t *testing.T) {
	s, _ := jsonvalue.DecodeLimited([]byte(`{"anyOf":[{"x-a":1}]}`), jsonvalue.NumberLimits{MaxDigits: 10, MaxExponent: 10})
	n, _ := compile.Compile(s)
	ev, err := EvaluateFull(n, true, DefaultMaxOutputUnits, DefaultMaxAnnotations)
	if err != nil || !ev.Valid || ev.Tree != nil || len(ev.Annotations) != 1 {
		t.Fatalf("got %+v %v", ev, err)
	}
}

func TestConcurrentEvaluateTree(t *testing.T) {
	want, err := evalFull(t, `{"properties":{"a":{"title":"A","anyOf":[{"type":"string"},{"x":1}]}}}`, `{"a":1}`, DefaultMaxAnnotations)
	if err != nil {
		t.Fatal(err)
	}
	s, _ := jsonvalue.DecodeLimited([]byte(`{"properties":{"a":{"title":"A","anyOf":[{"type":"string"},{"x":1}]}}}`), jsonvalue.NumberLimits{MaxDigits: 10, MaxExponent: 10})
	n, _ := compile.Compile(s)
	d, _ := jsonvalue.DecodeLimited([]byte(`{"a":1}`), jsonvalue.NumberLimits{MaxDigits: 10, MaxExponent: 10})
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				ev, err := EvaluateTree(n, d, DefaultMaxOutputUnits, DefaultMaxAnnotations)
				if err != nil || !reflect.DeepEqual(ev.Tree, want.Tree) || !reflect.DeepEqual(ev.Annotations, want.Annotations) {
					t.Errorf("differs: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func TestUnknownKeywordCollected(t *testing.T) {
	ev, err := evalFull(t, `{"x-note":"hi"}`, `1`, DefaultMaxAnnotations)
	if err != nil || !ev.Valid {
		t.Fatal(ev, err)
	}
	if len(ev.Annotations) != 1 || ev.Annotations[0].Keyword != "x-note" || ev.Annotations[0].Value != "hi" || ev.Annotations[0].KeywordLocation != "/x-note" {
		t.Fatalf("got %+v", ev.Annotations)
	}
}

func TestFailingBranchAnnotationsDropped(t *testing.T) {
	// the first anyOf branch fails after its sibling x-a annotated
	ev, err := evalFull(t, `{"anyOf":[{"x-a":1,"type":"string"},{"x-b":2}]}`, `5`, DefaultMaxAnnotations)
	if err != nil || !ev.Valid {
		t.Fatal(ev, err)
	}
	if len(ev.Annotations) != 1 || ev.Annotations[0].Keyword != "x-b" {
		t.Fatalf("got %+v", ev.Annotations)
	}
	anyOf := ev.Tree.Children
	if len(anyOf) != 2 || anyOf[0].Valid || anyOf[0].Annotations != nil || !anyOf[1].Valid || len(anyOf[1].Annotations) != 1 {
		t.Fatalf("tree wrong: %+v", anyOf)
	}
}

func TestFailingResultHasNoAnnotationsButTreeKeepsPassingPart(t *testing.T) {
	ev, err := evalFull(t, `{"allOf":[{"x-a":1},{"type":"string"}]}`, `5`, DefaultMaxAnnotations)
	if err != nil || ev.Valid {
		t.Fatal(ev, err)
	}
	if len(ev.Annotations) != 0 || ev.Tree.Annotations != nil {
		t.Fatalf("annotations leaked: %+v", ev.Annotations)
	}
	if len(ev.Tree.Children) != 2 || !ev.Tree.Children[0].Valid || ev.Tree.Children[1].Valid || len(ev.Tree.Children[1].Errors) == 0 {
		t.Fatalf("tree wrong: %+v", ev.Tree.Children)
	}
}

func TestAnnotationLimit(t *testing.T) {
	_, err := evalFull(t, `{"x-a":1,"x-b":2}`, `1`, 1)
	var le *jsonvalue.LimitError
	if !errors.As(err, &le) || le.Limit != "annotations" {
		t.Fatalf("want annotations LimitError, got %v", err)
	}
	// dropped annotations count too
	_, err = evalFull(t, `{"anyOf":[{"x-a":1,"type":"string"}],"x-b":1}`, `1`, 1)
	if !errors.As(err, &le) {
		t.Fatalf("dropped annotation not counted: %v", err)
	}
}

func TestSortOrder(t *testing.T) {
	in := []Annotation{
		{InstanceLocation: "/a/b", Keyword: "k", Seq: 1},
		{InstanceLocation: "/a~1b", Keyword: "k", Seq: 2},
		{InstanceLocation: "/a", KeywordLocation: "/z", Keyword: "k", Seq: 3},
		{InstanceLocation: "/a", KeywordLocation: "/y", Keyword: "k", Seq: 4},
		{InstanceLocation: "/a", KeywordLocation: "/y", AbsoluteKeywordLocation: "u", Keyword: "k", Seq: 5},
		{InstanceLocation: "/a", KeywordLocation: "/y", Keyword: "k", Seq: 0},
	}
	want := []int{0, 4, 5, 3, 1, 2}
	got := SortAnnotations(in)
	for i, a := range got {
		if a.Seq != want[i] {
			t.Fatalf("pos %d: got seq %d want %d (%+v)", i, a.Seq, want[i], got)
		}
	}
}
