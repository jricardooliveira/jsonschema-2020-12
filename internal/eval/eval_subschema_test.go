package eval

import "testing"

func TestChildFailureCarriesRootedLocation(t *testing.T) {
	n := mustCompile(t, map[string]any{"x-test-keep": []any{map[string]any{"const": false}}})
	_, units, err := EvaluateUnits(n, "x", DefaultMaxOutputUnits)
	if err != nil || len(units) != 2 || units[0].KeywordLocation != "/x-test-keep/0/const" {
		t.Fatalf("units=%+v err=%v", units, err)
	}
}
