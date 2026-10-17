package unevaluated_test

import (
	"reflect"
	"strings"
	"testing"
)

func TestUnevaluatedItemsOneErrorPerItem(t *testing.T) {
	ok, locs := run(t, `{"prefixItems":[true],"unevaluatedItems":false}`, `[1,"SENTINEL","SENTINEL"]`)
	if ok || !reflect.DeepEqual(locs, []string{"/1", "/2"}) {
		t.Errorf("ok=%v locs=%v", ok, locs)
	}
}

func TestUnevaluatedItemsCoveredBySiblings(t *testing.T) {
	for _, s := range []string{
		`{"unevaluatedItems":false,"items":true}`,
		`{"unevaluatedItems":false,"allOf":[{"prefixItems":[true,true]}]}`,
		`{"unevaluatedItems":false,"contains":{"type":"integer"}}`,
		`{"allOf":[{"unevaluatedItems":true}],"unevaluatedItems":false}`,
	} {
		if ok, _ := run(t, s, `[1,2]`); !ok {
			t.Errorf("want valid: %s", s)
		}
	}
}

// core 11.2: unevaluatedItems sees only its own schema object's adjacent
// keywords and their successful subschemas, never a sibling allOf branch.
func TestUnevaluatedItemsScopedToSchemaObject(t *testing.T) {
	for _, c := range []struct {
		schema string
		valid  bool
	}{
		{`{"allOf":[{"prefixItems":[true]},{"unevaluatedItems":false}]}`, false},
		{`{"allOf":[{"prefixItems":[true],"unevaluatedItems":false}]}`, true},
		{`{"items":true,"allOf":[{"unevaluatedItems":false}]}`, false},
		{`{"allOf":[{"allOf":[{"prefixItems":[true]}]},{"allOf":[{"unevaluatedItems":false}]}]}`, false},
		{`{"allOf":[{"allOf":[{"prefixItems":[true]}],"unevaluatedItems":false}]}`, true},
	} {
		if ok, _ := run(t, c.schema, `[1]`); ok != c.valid {
			t.Errorf("%s: valid=%v want %v", c.schema, ok, c.valid)
		}
	}
}

// core 10.2.1.4 and 7.7.1: a failing 'not' contributes nothing.
func TestUnevaluatedItemsFailingNotContributesNothing(t *testing.T) {
	ok, units := runUnits(t, `{"not":{"prefixItems":[true]},"unevaluatedItems":false}`, `["SENTINEL"]`)
	if ok {
		t.Fatal("want invalid")
	}
	for _, u := range units {
		if u.InstanceLocation == "/0" && strings.HasPrefix(u.KeywordLocation, "/unevaluatedItems") {
			return
		}
	}
	t.Errorf("no unevaluatedItems error at /0: %v", units)
}

func TestUnevaluatedItemsErrorLocations(t *testing.T) {
	assertErrorAt(t, `{"$id":"http://x/root","$ref":"#/$defs/c","$defs":{"c":{"unevaluatedItems":false}}}`, `["SENTINEL"]`,
		basicError{"/0", "/$ref/unevaluatedItems", "http://x/root#/$defs/c/unevaluatedItems"})
}

func TestUnevaluatedItemsContainsOnlyMatches(t *testing.T) {
	ok, locs := run(t, `{"contains":{"type":"integer"},"unevaluatedItems":false}`, `[1,"x"]`)
	if ok || !reflect.DeepEqual(locs, []string{"/1"}) {
		t.Errorf("ok=%v locs=%v", ok, locs)
	}
}

func TestUnevaluatedItemsFailedBranchAndNonArray(t *testing.T) {
	if ok, _ := run(t, `{"anyOf":[{"prefixItems":[true],"minItems":5}],"unevaluatedItems":false}`, `[1]`); ok {
		t.Error("failed branch must not cover")
	}
	if ok, _ := run(t, `{"unevaluatedItems":false}`, `{"a":1}`); !ok {
		t.Error("non-array ignored")
	}
}
