package validation_test

import (
	"strings"
	"testing"
)

// The testdata/validation/dependentrequired.json cases run through the root TestSpecCases.

func TestDependentRequiredMessageAndOrder(t *testing.T) {
	out := basicOut(t, `{"dependentRequired":{"b":["y","x"],"a":["z"]}}`, `{"a":"secret","b":1}`)
	z, y, x := strings.Index(out, "property: z"), strings.Index(out, "property: y"), strings.Index(out, "property: x")
	if z < 0 || y < 0 || x < 0 || !(z < y && y < x) {
		t.Errorf("want triggers sorted, names in schema order: %s", out)
	}
	if strings.Contains(out, "secret") {
		t.Errorf("leak: %s", out)
	}
}

func TestDependentRequiredDeterministic(t *testing.T) {
	s, d := `{"dependentRequired":{"b":["y"],"a":["z"],"c":["w"]}}`, `{"c":1,"a":1,"b":1}`
	if basicOut(t, s, d) != basicOut(t, s, d) {
		t.Error("output differs between runs")
	}
}
