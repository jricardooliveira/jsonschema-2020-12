package unevaluated_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/eval"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
	_ "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords/all"
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
	ok, units := runUnits(t, schema, doc)
	var locs []string
	for _, u := range units {
		locs = append(locs, u.InstanceLocation)
	}
	return ok, locs
}

// runUnits evaluates doc and returns its failure units, checking that none
// quotes the SENTINEL instance value.
func runUnits(t *testing.T, schema, doc string) (bool, []keywords.Unit) {
	t.Helper()
	n, err := compile.Compile(decode(t, schema))
	if err != nil {
		t.Fatal(err)
	}
	ok, units, err := eval.EvaluateUnits(n, decode(t, doc), 1000)
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range units {
		if strings.Contains(fmt.Sprint(u), "SENTINEL") {
			t.Errorf("unit leaks instance value: %v", u)
		}
	}
	return ok, units
}

// core 11.3: unevaluatedProperties sees only its own schema object's adjacent
// keywords and their successful subschemas, never a sibling allOf branch.
func TestUnevaluatedPropertiesScopedToSchemaObject(t *testing.T) {
	for _, c := range []struct {
		schema string
		valid  bool
	}{
		{`{"allOf":[{"properties":{"a":{}}},{"unevaluatedProperties":false}]}`, false},
		{`{"allOf":[{"properties":{"a":{}},"unevaluatedProperties":false}]}`, true},
		{`{"properties":{"a":{}},"allOf":[{"unevaluatedProperties":false}]}`, false},
		{`{"allOf":[{"allOf":[{"properties":{"a":{}}}]},{"allOf":[{"unevaluatedProperties":false}]}]}`, false},
		{`{"allOf":[{"allOf":[{"properties":{"a":{}}}],"unevaluatedProperties":false}]}`, true},
	} {
		if ok, _ := run(t, c.schema, `{"a":1}`); ok != c.valid {
			t.Errorf("%s: valid=%v want %v", c.schema, ok, c.valid)
		}
	}
}

// core 10.2.1.4 and 7.7.1: a failing 'not' (its subschema passed) still
// contributes nothing, so 'a' is reported as unevaluated.
func TestUnevaluatedPropertiesFailingNotContributesNothing(t *testing.T) {
	ok, units := runUnits(t, `{"not":{"properties":{"a":true}},"unevaluatedProperties":false}`, `{"a":"SENTINEL"}`)
	if ok {
		t.Fatal("want invalid")
	}
	for _, u := range units {
		if u.InstanceLocation == "/a" && strings.HasPrefix(u.KeywordLocation, "/unevaluatedProperties") {
			return
		}
	}
	t.Errorf("no unevaluatedProperties error at /a: %v", units)
}

// core 12.3: each error carries the keyword location and, through a $ref,
// the absolute keyword location.
func TestUnevaluatedPropertiesErrorLocations(t *testing.T) {
	assertErrorAt(t, `{"$id":"http://x/root","$ref":"#/$defs/c","$defs":{"c":{"unevaluatedProperties":false}}}`, `{"b":"SENTINEL"}`,
		basicError{"/b", "/$ref/unevaluatedProperties", "http://x/root#/$defs/c/unevaluatedProperties"})
}

type basicError struct {
	InstanceLocation        string `json:"instanceLocation"`
	KeywordLocation         string `json:"keywordLocation"`
	AbsoluteKeywordLocation string `json:"absoluteKeywordLocation"`
}

// assertErrorAt checks that the basic report of doc has want among its errors.
func assertErrorAt(t *testing.T, schema, doc string, want basicError) {
	t.Helper()
	s, err := jsonschema.Compile([]byte(schema))
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Validate([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	out, err := r.Output("basic")
	if err != nil {
		t.Fatal(err)
	}
	var rep struct{ Errors []basicError }
	if err := json.Unmarshal(out, &rep); err != nil {
		t.Fatal(err)
	}
	if r.Valid() || !slices.Contains(rep.Errors, want) {
		t.Errorf("want %+v in %s", want, out)
	}
}

// Neither the error report nor the annotation report quotes an instance value.
func TestUnevaluatedOutputHasNoInstanceValues(t *testing.T) {
	s, err := jsonschema.Compile([]byte(`{"properties":{"a":true},"prefixItems":[true],"unevaluatedProperties":{"type":"string"},"unevaluatedItems":{"type":"string"}}`))
	if err != nil {
		t.Fatal(err)
	}
	for _, doc := range []string{`{"a":"SENTINEL","b":"SENTINEL"}`, `{"b":12345678}`, `["SENTINEL","SENTINEL"]`, `[1,12345678]`} {
		r, err := s.Validate([]byte(doc))
		if err != nil {
			t.Fatal(err)
		}
		out, err := r.Output("basic")
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(out), "SENTINEL") || strings.Contains(string(out), "12345678") {
			t.Errorf("%s: output leaks a value: %s", doc, out)
		}
	}
}

// F5: the lookup of evaluated locations counts evaluation steps, so a wide
// document cannot make it quadratic without hitting the step limit.
func TestUnevaluatedLookupCountsSteps(t *testing.T) {
	assertLookupCountsSteps(t, `{"properties":{"a":true,"b":true}}`, `{"properties":{"a":true,"b":true},"unevaluatedProperties":false}`, `{"a":1,"b":1}`)
	assertLookupCountsSteps(t, `{"prefixItems":[true,true]}`, `{"prefixItems":[true,true],"unevaluatedItems":false}`, `[1,2]`)
}

// assertLookupCountsSteps finds the smallest step cap base passes with, and
// checks that adding the unevaluated keyword (which applies to nothing, so
// only looks up) no longer fits in it.
func assertLookupCountsSteps(t *testing.T, base, withUneval, doc string) {
	t.Helper()
	nb, err := compile.Compile(decode(t, base))
	if err != nil {
		t.Fatal(err)
	}
	nu, err := compile.Compile(decode(t, withUneval))
	if err != nil {
		t.Fatal(err)
	}
	d := decode(t, doc)
	steps := 0
	for ; steps < 100; steps++ {
		if _, _, err := eval.EvaluateLimited(nb, d, 1000, steps); err == nil {
			break
		}
	}
	_, _, err = eval.EvaluateLimited(nu, d, 1000, steps)
	if _, ok := err.(*jsonvalue.LimitError); !ok {
		t.Errorf("%s: lookup counted no step at cap %d (err=%v)", withUneval, steps, err)
	}
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
