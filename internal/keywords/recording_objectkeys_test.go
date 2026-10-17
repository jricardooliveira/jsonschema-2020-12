package keywords

import (
	"reflect"
	"testing"
)

type fakeRec struct{ on bool }

func (f fakeRec) Recording() bool { return f.on }

func TestRecordingFollowsItems(t *testing.T) {
	if !(&EvalContext{}).Recording() {
		t.Error("no Items: must record (fail-safe)")
	}
	if !(&EvalContext{Items: struct{}{}}).Recording() {
		t.Error("Items without Recording: must record")
	}
	if (&EvalContext{Items: fakeRec{false}}).Recording() {
		t.Error("want false")
	}
	if !(&EvalContext{Items: fakeRec{true}}).Recording() {
		t.Error("want true")
	}
}

func TestAnnotateNamesSameValueAsAnnotate(t *testing.T) {
	ctx := NewEvalContext(10, nil)
	names := []string{"a", "b"}
	if err := ctx.AnnotateNames(names); err != nil {
		t.Fatal(err)
	}
	if err := ctx.Annotate(append([]string(nil), names...)); err != nil {
		t.Fatal(err)
	}
	got := ctx.Annotations()
	if len(got) != 2 || !reflect.DeepEqual(got[0].Value, got[1].Value) {
		t.Errorf("got %v", got)
	}
}
