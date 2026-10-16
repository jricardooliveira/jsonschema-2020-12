package core_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/neoface/jsonschema"
	"github.com/neoface/jsonschema/internal/eval"
)

// Valid/invalid cases live in testdata/core/dynamicRef.json and
// dynamicAnchor.json (run by the root TestSpecCases). The cases below need
// compile errors, a registry, a timeout or the report, which the JSON
// runner cannot express.

// core 8.2.2: duplicate plain names in one resource are a schema error.
func TestSpecDuplicatePlainNameIsSchemaError(t *testing.T) {
	for name, src := range map[string]string{
		"dynamic+dynamic": `{"$defs":{"a":{"$dynamicAnchor":"n"},"b":{"$dynamicAnchor":"n"}}}`,
		"dynamic+anchor":  `{"$defs":{"a":{"$dynamicAnchor":"n"},"b":{"$anchor":"n"}}}`,
	} {
		if _, err := jsonschema.Compile([]byte(src)); err == nil {
			t.Errorf("%s: want schema error", name)
		}
	}
}

// core 8.2.3.2 + 9.1.1 (offline): an unresolvable initial target is an error.
func TestSpecUnresolvedTargetIsSchemaError(t *testing.T) {
	for _, src := range []string{
		`{"$dynamicRef":"#nope"}`,
		`{"$dynamicRef":"http://x/missing#n"}`,
		`{"$dynamicRef":"file:///etc/passwd#n"}`,
	} {
		if _, err := jsonschema.Compile([]byte(src)); err == nil {
			t.Errorf("%s: want error", src)
		}
	}
}

// core 8.2.3.2 across documents: the dynamic scope spans registry resources.
func TestSpecRegistryDynamicScope(t *testing.T) {
	dir := t.TempDir()
	list := filepath.Join(dir, "list.json")
	if err := os.WriteFile(list, []byte(`{"$id":"http://x/list","type":"array","items":{"$dynamicRef":"#T"},
 "$defs":{"default":{"$dynamicAnchor":"T","type":"number"}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := jsonschema.CompileWithRegistry([]byte(`{"$id":"http://x/root","$dynamicAnchor":"T","$ref":"http://x/list"}`),
		jsonschema.Options{}, jsonschema.Registry{Files: []jsonschema.RegistryFile{{Path: list}}})
	if err != nil {
		t.Fatal(err)
	}
	for doc, want := range map[string]bool{`[]`: true, `[[]]`: true, `[1]`: false} {
		r, err := s.Validate([]byte(doc))
		if err != nil || r.Valid() != want {
			t.Errorf("%s: valid=%v err=%v", doc, r.Valid(), err)
		}
	}
}

// core 8.2.3.1/8.2.3.2: a cycle that never descends ends with a loop error.
func TestSpecCycleIsErrorNotHang(t *testing.T) {
	s, err := jsonschema.Compile([]byte(`{"$defs":{"a":{"$dynamicAnchor":"a","$dynamicRef":"#a"}},"$ref":"#/$defs/a"}`))
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := s.Validate([]byte(`{}`)); done <- err }()
	select {
	case err := <-done:
		var le *eval.LoopError
		if err == nil || !errors.As(err, &le) {
			t.Errorf("want loop error, got %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("cycle did not terminate")
	}
}

// A legitimate re-entry of the same resource at a deeper instance location
// is not a loop.
func TestSpecRecursionIntoDocumentIsNotLoop(t *testing.T) {
	s, err := jsonschema.Compile([]byte(`{"$id":"https://t.example/t","$dynamicAnchor":"n","properties":{"c":{"$dynamicRef":"#n"}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if r, err := s.Validate([]byte(`{"c":{"c":{"c":{}}}}`)); err != nil || !r.Valid() {
		t.Errorf("valid=%v err=%v", r.Valid(), err)
	}
}

// core 8.2.3.2 / output 12.3: error location names /$dynamicRef, and the
// report never quotes the instance value.
func TestSpecErrorLocationAndNoInstanceValue(t *testing.T) {
	s, err := jsonschema.Compile([]byte(`{"$defs":{"a":{"$anchor":"a","type":"number"}},"$dynamicRef":"#a"}`))
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Validate([]byte(`"secret-value-123"`))
	if err != nil {
		t.Fatal(err)
	}
	if r.Valid() {
		t.Fatal("want invalid")
	}
	out, err := r.Output("basic")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "/$dynamicRef") {
		t.Errorf("no /$dynamicRef in %s", out)
	}
	if strings.Contains(string(out), "secret-value-123") {
		t.Errorf("instance value leaked: %s", out)
	}
}
