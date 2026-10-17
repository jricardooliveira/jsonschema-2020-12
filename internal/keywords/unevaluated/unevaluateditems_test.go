package unevaluated_test

import (
	"reflect"
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
