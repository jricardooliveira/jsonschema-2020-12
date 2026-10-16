package registry

import (
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/jsonvalue"
)

const root = "file:///s.json"

func build(t *testing.T, src string) (*Index, error) {
	t.Helper()
	doc, err := jsonvalue.Decode([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	return Build(doc, root)
}

func TestEmbeddedResourcesAndBases(t *testing.T) { // core 8.2.1, 9.2
	idx, err := build(t, `{"$id":"https://Example.com/a/root.json","$defs":{
		"e":{"$id":"sub/e.json","$anchor":"x","properties":{"p":{}}}},"$anchor":"x"}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(idx.Resources) != 2 {
		t.Fatalf("resources: %v", idx.Resources)
	}
	e := idx.Resources["https://example.com/a/sub/e.json"]
	if e == nil || e.Pointer != "/$defs/e" || e.Anchors["x"] != "/$defs/e" {
		t.Fatalf("embedded: %+v", e)
	}
	if r := idx.Resources["https://example.com/a/root.json"]; r == nil || r.Anchors["x"] != "" {
		t.Fatalf("root: %+v", r)
	}
	if got := idx.Bases["/$defs/e/properties/p"]; got != "https://example.com/a/sub/e.json" {
		t.Errorf("base = %q", got)
	}
}

func TestRootWithoutID(t *testing.T) { // core 9.1
	idx, err := build(t, `{"$defs":{"a":{"$anchor":"A"}}}`)
	if err != nil {
		t.Fatal(err)
	}
	if idx.Resources[root].Anchors["A"] != "/$defs/a" || idx.Bases["/$defs/a"] != root {
		t.Errorf("%+v", idx)
	}
}

func TestIDFragment(t *testing.T) { // core 8.2.1
	idx, err := build(t, `{"$id":"http://x/y#"}`)
	if err != nil || idx.Resources["http://x/y"] == nil {
		t.Fatalf("empty fragment: %v %v", idx, err)
	}
	_, err = build(t, `{"$defs":{"a":{"$id":"http://x/y#frag"}}}`)
	if err == nil || !strings.Contains(err.Error(), "/$defs/a") {
		t.Errorf("non-empty fragment: %v", err)
	}
}

func TestAnchorSyntax(t *testing.T) { // core 8.2.2
	for _, ok := range []string{"a", "_a", "A-b.c_1"} {
		if _, err := build(t, `{"$anchor":"`+ok+`"}`); err != nil {
			t.Errorf("%q: %v", ok, err)
		}
	}
	for _, bad := range []string{"1a", "-a", "a b", "a#", ""} {
		for _, kw := range []string{"$anchor", "$dynamicAnchor"} {
			if _, err := build(t, `{"`+kw+`":"`+bad+`"}`); err == nil {
				t.Errorf("%s %q accepted", kw, bad)
			}
		}
	}
}

func TestDuplicates(t *testing.T) { // core 8.2.1, 8.2.2
	_, err := build(t, `{"allOf":[{"$id":"http://x/a"},{"$id":"http://X/a"}]}`)
	if err == nil || !strings.Contains(err.Error(), "http://x/a") {
		t.Errorf("dup $id: %v", err)
	}
	_, err = build(t, `{"allOf":[{"$anchor":"q"},{"$dynamicAnchor":"q"}]}`)
	if err == nil || !strings.Contains(err.Error(), `"q"`) {
		t.Errorf("dup anchor: %v", err)
	}
	// same anchor name in different resources is fine
	if _, err := build(t, `{"$anchor":"q","allOf":[{"$id":"http://x/a","$anchor":"q"}]}`); err != nil {
		t.Errorf("distinct resources: %v", err)
	}
}

func TestNonSchemaValuesNotWalked(t *testing.T) {
	// $id inside enum/const is data, not a resource.
	idx, err := build(t, `{"enum":[{"$id":"http://x/a#f"}],"const":{"$anchor":"1"}}`)
	if err != nil || len(idx.Resources) != 1 {
		t.Errorf("%v %v", idx, err)
	}
}

func TestPointerEscapeInBases(t *testing.T) {
	idx, err := build(t, `{"$defs":{"a/b~c":{"$id":"http://x/z"}}}`)
	if err != nil || idx.Resources["http://x/z"].Pointer != "/$defs/a~1b~0c" {
		t.Errorf("%v %v", idx, err)
	}
}
