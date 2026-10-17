package unevaluated_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/eval"
	"github.com/neoface/jsonschema/internal/jsonvalue"
	_ "github.com/neoface/jsonschema/internal/keywords/all"
)

func decode(t *testing.T, s string) any {
	t.Helper()
	v, err := jsonvalue.DecodeLimited([]byte(s), jsonvalue.NumberLimits{})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func run(t *testing.T, schema, doc string) (bool, []string) {
	t.Helper()
	n, err := compile.Compile(decode(t, schema))
	if err != nil {
		t.Fatal(err)
	}
	ok, units, err := eval.EvaluateUnits(n, decode(t, doc), 1000)
	if err != nil {
		t.Fatal(err)
	}
	var locs []string
	for _, u := range units {
		locs = append(locs, u.InstanceLocation)
		if strings.Contains(fmt.Sprint(u), "SENTINEL") {
			t.Errorf("unit leaks instance value: %v", u)
		}
	}
	return ok, locs
}

func TestUnevaluatedPropertiesOneErrorPerProperty(t *testing.T) {
	ok, locs := run(t, `{"properties":{"a":true},"unevaluatedProperties":false}`,
		`{"a":1,"nickname":"SENTINEL","age":"SENTINEL","x/y":1}`)
	if ok {
		t.Fatal("want invalid")
	}
	want := []string{"/age", "/nickname", "/x~1y"}
	if !reflect.DeepEqual(locs, want) {
		t.Errorf("got %v want %v", locs, want)
	}
}

func TestUnevaluatedPropertiesCoveredBySiblingsWhateverOrder(t *testing.T) {
	ok, _ := run(t, `{"unevaluatedProperties":false,"allOf":[{"properties":{"a":true}}],"patternProperties":{"^p":true}}`, `{"a":1,"pz":2}`)
	if !ok {
		t.Fatal("want valid")
	}
}

func TestUnevaluatedPropertiesFailedBranchCountsForNothing(t *testing.T) {
	ok, locs := run(t, `{"anyOf":[{"properties":{"a":true},"required":["zz"]}],"unevaluatedProperties":false}`, `{"a":1}`)
	if ok {
		t.Fatal("want invalid")
	}
	_ = locs
	ok, _ = run(t, `{"not":{"properties":{"a":true},"required":["q"]},"unevaluatedProperties":false}`, `{"a":1}`)
	if ok {
		t.Fatal("not must not cover")
	}
}

func TestUnevaluatedPropertiesNestedAndNonObject(t *testing.T) {
	ok, _ := run(t, `{"allOf":[{"unevaluatedProperties":true}],"unevaluatedProperties":false}`, `{"a":1}`)
	if !ok {
		t.Fatal("nested unevaluated must cover")
	}
	ok, _ = run(t, `{"unevaluatedProperties":false}`, `[1]`)
	if !ok {
		t.Fatal("non-object ignored")
	}
}

func TestUnevaluatedPropertiesSubschemaApplied(t *testing.T) {
	ok, locs := run(t, `{"unevaluatedProperties":{"type":"integer"}}`, `{"a":1,"b":"SENTINEL"}`)
	if ok || !reflect.DeepEqual(locs, []string{"/b"}) {
		t.Errorf("ok=%v locs=%v", ok, locs)
	}
}
