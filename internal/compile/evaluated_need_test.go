package compile

import (
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/registry"
)

func needs(t *testing.T, schema string, ext *registry.Index) bool {
	t.Helper()
	v, err := jsonvalue.Decode([]byte(schema))
	if err != nil {
		t.Fatal(err)
	}
	return NeedsEvaluated([]Doc{{Value: v}}, ext)
}

func TestNeedsEvaluated(t *testing.T) {
	ext, err := registry.Build(true, "https://x.test/base")
	if err != nil {
		t.Fatal(err)
	}
	if err := ext.AddBuiltins(); err != nil {
		t.Fatal(err)
	}
	other := `{"$id":"https://x.test/other","$dynamicAnchor":"n","unevaluatedProperties":false}`
	ov, _ := jsonvalue.Decode([]byte(other))
	if err := ext.AddExternal(ov, "https://x.test/other", "other.json"); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, schema string
		ext          *registry.Index
		want         bool
	}{
		{"none", `{"type":"object","properties":{"a":{"type":"string"}}}`, nil, false},
		{"direct", `{"unevaluatedProperties":false}`, nil, true},
		{"direct items nested", `{"properties":{"a":{"unevaluatedItems":false}}}`, nil, true},
		{"behind $ref", `{"$ref":"#/$defs/a","$defs":{"a":{"unevaluatedItems":false}}}`, nil, true},
		{"registry via $dynamicRef", `{"$dynamicRef":"https://x.test/other#n"}`, ext, true},
		{"cycle of refs terminates", `{"$ref":"#/$defs/a","$defs":{"a":{"$ref":"#/$defs/b"},"b":{"$ref":"#/$defs/a"}}}`, nil, false},
		{"boolean subschemas", `{"properties":{"a":true,"b":false},"items":true}`, nil, false},
		{"boolean root", `true`, nil, false},
		{"external ref, unscannable registry", `{"$ref":"https://x.test/other"}`, nil, true},
	}
	for _, c := range cases {
		if got := needs(t, c.schema, c.ext); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
