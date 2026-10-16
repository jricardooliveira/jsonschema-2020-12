package jsonschema

import "testing"

const relRefSchema = `{"$defs":{"e":{"$id":"http://x/y/e.json"}},"$ref":"e.json"}`

// core 9.1: the base URI decides how a relative $ref resolves.
func TestBaseURIOption(t *testing.T) {
	if _, err := CompileWithOptions([]byte(relRefSchema), Options{BaseURI: "http://x/y/s.json"}); err != nil {
		t.Errorf("override: %v", err)
	}
	if _, err := CompileWithOptions([]byte(relRefSchema), Options{BaseURI: "http://x/z/s.json"}); err == nil {
		t.Error("expected unresolved reference")
	}
	if _, err := Compile([]byte(relRefSchema)); err == nil {
		t.Error("default (stdin placeholder) base must not resolve e.json")
	}
}

func TestBaseURIOptionRejectsBadValues(t *testing.T) {
	for _, bad := range []string{"rel/path", "http://x/a#frag", "http://x/%zz"} {
		if _, err := CompileWithOptions([]byte(`true`), Options{BaseURI: bad}); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestSchemaPathGivesFileBase(t *testing.T) {
	s := `{"$id":"file:///d/e.json","$ref":"file:///d/e.json#/nope"}`
	if _, err := CompileWithOptions([]byte(s), Options{SchemaPath: "/d/s.json"}); err == nil {
		t.Error("expected error")
	}
	rel := `{"$defs":{"e":{"$id":"file:///d/e.json"}},"$ref":"e.json"}`
	if _, err := CompileWithOptions([]byte(rel), Options{SchemaPath: "/d/s.json"}); err != nil {
		t.Error(err)
	}
}
