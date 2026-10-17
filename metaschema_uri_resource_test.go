package jsonschema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A caller resource at the built-in 2020-12 meta-schema address is accepted
// when JSON-equal to the embedded one and refused otherwise.
func TestCallerResourceAtBuiltinMetaSchemaURI(t *testing.T) {
	embedded, err := os.ReadFile("internal/registry/metaschemas/schema.json")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	write := func(name, content string) Registry {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		return Registry{Files: []RegistryFile{{Path: p}}}
	}

	// Same document, different layout: JSON-equal.
	same := strings.Replace(string(embedded), "{", "{ ", 1)
	for i := 0; i < 2; i++ { // twice: the second Compile uses the shared cache
		s, err := CompileWithRegistry([]byte(`{"type":"integer"}`), Options{}, write("same.json", same))
		if err != nil {
			t.Fatalf("JSON-equal copy: %v", err)
		}
		if !mustValidate(t, s, `1`) || mustValidate(t, s, `"x"`) {
			t.Fatal("wrong verdicts")
		}
	}

	different := strings.Replace(string(embedded), "{", `{"x-extra":true,`, 1)
	_, err = CompileWithRegistry([]byte(`true`), Options{}, write("diff.json", different))
	if err == nil || !strings.Contains(err.Error(), "duplicate address") ||
		!strings.Contains(err.Error(), "https://json-schema.org/draft/2020-12/schema") {
		t.Fatalf("different document: got %v, want the duplicate-address error", err)
	}
}
