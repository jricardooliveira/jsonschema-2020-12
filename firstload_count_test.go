package jsonschema

import (
	"testing"

	"github.com/neoface/jsonschema/internal/registry"
)

// TestFirstLoadBuildsBuiltinIndexOnce checks, in a fresh process, that the
// first public Compile (AddBuiltins plus the meta-schema cache) builds the
// built-in index exactly once. The count is exact, so race mode needs no
// separate limit.
func TestFirstLoadBuildsBuiltinIndexOnce(t *testing.T) {
	if _, child := freshProcess(t); !child {
		return
	}
	if _, err := Compile([]byte(`{"type":"string"}`)); err != nil {
		t.Fatal(err)
	}
	if n := registry.BuiltinBuilds(); n != 1 {
		t.Fatalf("built-in index builds = %d, want 1", n)
	}
}
