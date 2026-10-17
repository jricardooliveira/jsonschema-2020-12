package jsonschema

import (
	"encoding/json"
	"testing"
)

// Nested $ref, if and allOf in verbose output: frames are reused across depths
// and siblings, so every unit must still carry its own instance, keyword and
// absolute keyword location.
func TestVerboseNestedRefIfAllOfLocations(t *testing.T) {
	s, err := Compile([]byte(`{
		"$id": "https://example.com/root.json",
		"$defs": {"n": {"allOf": [
			{"if": {"type": "integer"}, "then": {"minimum": 5}},
			{"type": ["integer", "string"]}
		]}},
		"properties": {"a": {"$ref": "#/$defs/n"}, "b": {"$ref": "#/$defs/n"}}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	res, err := s.Validate([]byte(`{"a": 1, "b": 7}`))
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
	var found bool
	var walk func(u map[string]any)
	walk = func(u map[string]any) {
		if u["error"] != nil && u["instanceLocation"] == "/a" &&
			u["keywordLocation"] == "/properties/a/$ref/allOf/0/then/minimum" {
			found = true
			if got := u["absoluteKeywordLocation"]; got != "https://example.com/root.json#/$defs/n/allOf/0/then/minimum" {
				t.Errorf("absoluteKeywordLocation = %v", got)
			}
		}
		for _, k := range []string{"errors", "annotations"} {
			if l, ok := u[k].([]any); ok {
				for _, c := range l {
					walk(c.(map[string]any))
				}
			}
		}
	}
	walk(out)
	if !found {
		t.Errorf("minimum error at /a not found in %s", raw)
	}
	t.Logf("%s", raw)
}
