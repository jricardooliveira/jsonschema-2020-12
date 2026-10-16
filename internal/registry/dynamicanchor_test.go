package registry

import "testing"

func TestDynamicAnchorIndexed(t *testing.T) { // core 8.2.2
	idx, err := build(t, `{"$defs":{"a":{"$dynamicAnchor":"n"},"b":{"$anchor":"s"}}}`)
	if err != nil {
		t.Fatal(err)
	}
	r := idx.Resources[root]
	if r.Anchors["n"] != "/$defs/a" {
		t.Errorf("#n should resolve to the dynamic anchor: %v", r.Anchors)
	}
	if !r.HasDynamicAnchor("n") {
		t.Error("n should be dynamic")
	}
	if r.HasDynamicAnchor("s") || r.HasDynamicAnchor("missing") {
		t.Error("static or missing anchor reported as dynamic")
	}
}

func TestDuplicatePlainNameRejected(t *testing.T) { // core 8.2.2
	for name, src := range map[string]string{
		"dynamic+dynamic": `{"$defs":{"a":{"$dynamicAnchor":"n"},"b":{"$dynamicAnchor":"n"}}}`,
		"dynamic+static":  `{"$defs":{"a":{"$dynamicAnchor":"n"},"b":{"$anchor":"n"}}}`,
		"static+dynamic":  `{"$defs":{"a":{"$anchor":"n"},"b":{"$dynamicAnchor":"n"}}}`,
	} {
		if _, err := build(t, src); err == nil {
			t.Errorf("%s: want duplicate error", name)
		}
	}
}

func TestSameNameInSeparateResourcesAllowed(t *testing.T) { // core 8.2.1
	if _, err := build(t, `{"$dynamicAnchor":"n","$defs":{"e":{"$id":"e.json","$dynamicAnchor":"n"}}}`); err != nil {
		t.Fatal(err)
	}
}
