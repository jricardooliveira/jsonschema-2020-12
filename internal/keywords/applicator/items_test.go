package applicator_test

import (
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/eval"
)

// The testdata/applicator/items.json cases run through the root TestSpecCases.

func TestItemsAnnotation(t *testing.T) {
	n := mustCompile(t, `{"prefixItems":[true],"items":true}`)
	for doc, want := range map[string]bool{`[1,2]`: true, `[1]`: false, `[]`: false, `5`: false} {
		ok, rec, err := eval.EvaluateAnnotated(n, mustDecode(t, doc))
		if err != nil || !ok {
			t.Fatalf("%s: ok=%v err=%v", doc, ok, err)
		}
		a, got := rec.At("")["/items"]
		if got != want || (got && !a.All) {
			t.Errorf("%s: recorded=%v %+v, want %v", doc, got, a, want)
		}
	}
}

func TestItemsReportsEveryFailingPosition(t *testing.T) {
	n := mustCompile(t, `{"prefixItems":[true],"items":{"type":"integer"}}`)
	ok, units, err := eval.EvaluateUnits(n, mustDecode(t, `["a","s1",2,"s2","s3"]`), 100)
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	var locs []string
	for _, u := range units {
		locs = append(locs, u.InstanceLocation+" "+u.KeywordLocation)
		if strings.Contains(u.Error, "s1") {
			t.Errorf("leaks value: %q", u.Error)
		}
	}
	want := "/1 /items/type,/3 /items/type,/4 /items/type" // /2 is an integer
	if strings.Join(locs, ",") != want {
		t.Errorf("got %v", locs)
	}
}

func TestItemsFalseUnitLocation(t *testing.T) {
	n := mustCompile(t, `{"prefixItems":[true],"items":false}`)
	_, units, _ := eval.EvaluateUnits(n, mustDecode(t, `[1,2]`), 100)
	if len(units) != 1 || units[0].InstanceLocation != "/1" || units[0].KeywordLocation != "/items" {
		t.Errorf("units=%+v", units)
	}
}

func TestItemsSchemaError(t *testing.T) {
	v := mustDecode(t, `{"items":"x"}`)
	if _, err := compile.Compile(v); err == nil {
		t.Error("want schema error")
	}
}
