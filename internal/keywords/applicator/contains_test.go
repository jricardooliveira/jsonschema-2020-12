package applicator_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/eval"
)

// The testdata/applicator/contains.json cases run through the root TestSpecCases.

func TestContainsAnnotation(t *testing.T) {
	n := mustCompile(t, `{"contains":{"type":"integer"}}`)
	cases := []struct {
		doc     string
		applied bool
		all     bool
		idx     []int
	}{
		{`[1,"a",2]`, true, false, []int{0, 2}},
		{`[1,2]`, true, true, nil},
		{`"x"`, false, false, nil},
	}
	for _, c := range cases {
		ok, rec, err := eval.EvaluateAnnotated(n, mustDecode(t, c.doc))
		if err != nil || !ok {
			t.Fatalf("%s: ok=%v err=%v", c.doc, ok, err)
		}
		a, got := rec.At("")["/contains"]
		if got != c.applied || a.All != c.all || !reflect.DeepEqual(a.Indices, c.idx) {
			t.Errorf("%s: %+v recorded=%v", c.doc, a, got)
		}
	}
}

func TestContainsSingleErrorNoValues(t *testing.T) {
	n := mustCompile(t, `{"contains":{"type":"integer"},"minContains":3}`)
	ok, units, err := eval.EvaluateUnits(n, mustDecode(t, `[1,"secret",2]`), 100)
	if err != nil || ok || len(units) != 1 {
		t.Fatalf("ok=%v err=%v units=%v", ok, err, units)
	}
	u := units[0]
	if u.InstanceLocation != "" || strings.Contains(u.Error, "secret") ||
		!strings.Contains(u.Error, "2") || !strings.Contains(u.Error, "3") || !strings.Contains(u.Error, "minContains") {
		t.Errorf("unit=%+v", u)
	}
	n = mustCompile(t, `{"contains":true,"maxContains":1}`)
	_, units, _ = eval.EvaluateUnits(n, mustDecode(t, `[1,2]`), 100)
	if len(units) != 1 || !strings.Contains(units[0].Error, "maxContains") {
		t.Errorf("units=%v", units)
	}
}

func TestContainsFailedRecordsNothingAndCapCounts(t *testing.T) {
	n := mustCompile(t, `{"contains":false}`)
	ok, rec, err := eval.EvaluateAnnotated(n, mustDecode(t, `[1]`))
	if err != nil || ok || len(rec.At("")) != 0 {
		t.Errorf("ok=%v err=%v", ok, err)
	}
	// discarded child units still count against the cap
	_, _, err = eval.EvaluateUnits(n, mustDecode(t, `[1,2,3]`), 2)
	if err == nil {
		t.Error("want LimitError")
	}
}

func TestContainsSchemaErrors(t *testing.T) {
	for _, s := range []string{`{"contains":"x"}`, `{"contains":true,"minContains":"x"}`, `{"contains":true,"maxContains":null}`} {
		if _, err := compile.Compile(mustDecode(t, s)); err == nil {
			t.Errorf("%s: want error", s)
		}
	}
}
