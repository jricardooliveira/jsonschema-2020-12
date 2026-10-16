package applicator_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/eval"
)

// The testdata/applicator/properties.json cases run through the root TestSpecCases.

func TestPropertiesKeyOnlyInLocation(t *testing.T) {
	n := mustCompile(t, `{"properties":{"a~b/c@x.com":{"type":"integer"}}}`)
	ok, units, err := eval.EvaluateUnits(n, mustDecode(t, `{"a~b/c@x.com":"v"}`), 100)
	if err != nil || ok || len(units) == 0 {
		t.Fatalf("ok=%v err=%v units=%+v", ok, err, units)
	}
	for _, u := range units {
		if u.InstanceLocation != "/a~0b~1c@x.com" {
			t.Errorf("location %q", u.InstanceLocation)
		}
		if strings.Contains(u.Error, "x.com") {
			t.Errorf("message leaks key: %q", u.Error)
		}
	}
}

func TestPropertiesRecordsAppliedNames(t *testing.T) {
	n := mustCompile(t, `{"properties":{"b":true,"a":false,"z":true}}`)
	_, rec, err := eval.EvaluateAnnotated(n, mustDecode(t, `{"b":1,"a":1,"q":1}`))
	if err != nil {
		t.Fatal(err)
	}
	_ = rec // failed root rolls back; use a passing schema below
	n = mustCompile(t, `{"properties":{"b":true,"a":true,"z":true}}`)
	_, rec, _ = eval.EvaluateAnnotated(n, mustDecode(t, `{"b":1,"a":1,"q":1}`))
	got := rec.PropertiesAt("")["/properties"]
	if !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Errorf("got %v", got)
	}
}

// A failing subschema evaluated through EvalChild drops the property names
// it recorded; a passing sibling keeps them. contains passes overall, so the
// root does not roll everything back.
func TestPropertiesRolledBackWhenSubschemaFails(t *testing.T) {
	n := mustCompile(t, `{"contains":{"properties":{"a":true},"required":["z"]}}`)
	ok, rec, err := eval.EvaluateAnnotated(n, mustDecode(t, `[{"a":1},{"a":1,"z":1}]`))
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if got := rec.PropertiesAt("/0"); len(got) != 0 {
		t.Errorf("failed element kept property annotations: %v", got)
	}
	if got := rec.PropertiesAt("/1")["/contains/properties"]; !reflect.DeepEqual(got, []string{"a"}) {
		t.Errorf("passing element: got %v", rec.PropertiesAt("/1"))
	}
}

func TestObjectOutputDeterministicAndOrdered(t *testing.T) {
	n := mustCompile(t, `{"properties":{"b":false,"a":false},"additionalProperties":false}`)
	var doc strings.Builder
	doc.WriteString(`{"b":1,"a":1`)
	for _, k := range []string{"y", "x", "w", "v", "u", "t", "s", "r", "q", "p", "o", "n", "m", "l", "k", "j", "i", "h", "g", "f", "e"} {
		doc.WriteString(`,"` + k + `":1`)
	}
	doc.WriteString("}")
	inst := mustDecode(t, doc.String())
	_, first, err := eval.EvaluateUnits(n, inst, 1000)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		_, again, _ := eval.EvaluateUnits(n, inst, 1000)
		if !reflect.DeepEqual(first, again) {
			t.Fatal("units differ between runs")
		}
	}
	// additionalProperties sorts before properties; members sorted within each.
	var locs []string
	for _, u := range first {
		if strings.HasPrefix(u.KeywordLocation, "/additionalProperties") {
			locs = append(locs, u.InstanceLocation)
		}
	}
	if len(locs) != 21 || locs[0] != "/e" || locs[20] != "/y" {
		t.Errorf("additionalProperties order: %v", locs)
	}
	lastAdd, firstProp := -1, -1
	for i, u := range first {
		if strings.HasPrefix(u.KeywordLocation, "/additionalProperties") {
			lastAdd = i
		} else if strings.HasPrefix(u.KeywordLocation, "/properties") && firstProp < 0 {
			firstProp = i
		}
	}
	if lastAdd > firstProp {
		t.Errorf("keyword order wrong: lastAdd=%d firstProp=%d", lastAdd, firstProp)
	}
}

func TestPropertiesLimit(t *testing.T) {
	n := mustCompile(t, `{"properties":{"a":false,"b":false,"c":false}}`)
	_, _, err := eval.EvaluateUnits(n, mustDecode(t, `{"a":1,"b":1,"c":1}`), 2)
	if err == nil {
		t.Error("want limit error")
	}
}
