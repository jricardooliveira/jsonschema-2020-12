package jsonschema

import (
	"encoding/json"
	"slices"
	"testing"
)

const nestedFramesSchema = `{
	"$id": "https://example.com/root.json",
	"$defs": {"n": {"allOf": [
		{"if": {"type": "integer"}, "then": {"minimum": 5}},
		{"type": ["integer", "string"]}
	]}},
	"properties": {"a": {"$ref": "#/$defs/n"}, "b": {"$ref": "#/$defs/n"}}
}`

// verboseUnits validates doc against the nested schema and returns every
// verbose output unit, depth first.
func verboseUnits(t *testing.T, doc string) []map[string]any {
	t.Helper()
	s, err := Compile([]byte(nestedFramesSchema))
	if err != nil {
		t.Fatal(err)
	}
	res, err := s.Validate([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	if res.Valid() {
		t.Fatal("want invalid")
	}
	raw, err := res.Output("verbose")
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	var units []map[string]any
	var walk func(u map[string]any)
	walk = func(u map[string]any) {
		units = append(units, u)
		for _, k := range []string{"errors", "annotations"} {
			if l, ok := u[k].([]any); ok {
				for _, c := range l {
					walk(c.(map[string]any))
				}
			}
		}
	}
	walk(out)
	return units
}

func unitsAt(units []map[string]any, inst string) []map[string]any {
	var at []map[string]any
	for _, u := range units {
		if u["instanceLocation"] == inst {
			at = append(at, u)
		}
	}
	return at
}

func minimumErrorAt(t *testing.T, units []map[string]any, prop string) {
	t.Helper()
	for _, u := range unitsAt(units, "/"+prop) {
		if u["error"] != nil && u["keywordLocation"] == "/properties/"+prop+"/$ref/allOf/0/then/minimum" {
			if got := u["absoluteKeywordLocation"]; got != "https://example.com/root.json#/$defs/n/allOf/0/then/minimum" {
				t.Errorf("/%s: absoluteKeywordLocation = %v", prop, got)
			}
			return
		}
	}
	t.Errorf("minimum error at /%s not found", prop)
}

// Nested $ref, if and allOf in verbose output: frames are reused across depths
// and siblings, so every unit must still carry its own instance, keyword and
// absolute keyword location.
func TestVerboseNestedRefIfAllOfLocations(t *testing.T) {
	units := verboseUnits(t, `{"a": 1, "b": 7}`)
	minimumErrorAt(t, units, "a")

	// Sibling /b is evaluated after /a left the same depth frames: each of
	// its units must name /b's own path, never a stale /a location.
	var got []string
	for _, u := range unitsAt(units, "/b") {
		got = append(got, u["keywordLocation"].(string))
		if u["valid"] != true {
			t.Errorf("/b unit %v not valid", u["keywordLocation"])
		}
		if abs, ok := u["absoluteKeywordLocation"]; ok {
			t.Errorf("valid /b unit %v has absoluteKeywordLocation %v", u["keywordLocation"], abs)
		}
	}
	want := []string{
		"/properties/b",
		"/properties/b/$ref",
		"/properties/b/$ref/allOf/0",
		"/properties/b/$ref/allOf/0/if",
		"/properties/b/$ref/allOf/0/then",
		"/properties/b/$ref/allOf/1",
	}
	if !slices.Equal(got, want) {
		t.Errorf("/b keyword locations = %q, want %q", got, want)
	}
}

// Both siblings fail: the second one's error carries its own keyword and
// absolute keyword locations through the reused frames.
func TestVerboseNestedRefIfAllOfSiblingErrors(t *testing.T) {
	units := verboseUnits(t, `{"a": 1, "b": 2}`)
	minimumErrorAt(t, units, "a")
	minimumErrorAt(t, units, "b")
}
