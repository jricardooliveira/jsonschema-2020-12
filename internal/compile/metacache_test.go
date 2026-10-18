package compile

import (
	"sync"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/registry"
)

func TestBuiltinSetBuiltOnce(t *testing.T) {
	var wg sync.WaitGroup
	sets := make([]*builtinSet, 16)
	for i := range sets {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s, err := sharedBuiltin()
			if err != nil {
				t.Error(err)
			}
			sets[i] = s
		}()
	}
	wg.Wait()
	for _, s := range sets {
		if s == nil || s != sets[0] {
			t.Fatal("shared set differs between calls")
		}
	}
	calls := 0
	for i := 0; i < 5; i++ {
		ext := &registry.Index{Resources: map[string]*registry.Resource{}}
		if err := ext.AddBuiltins(); err != nil {
			t.Fatal(err)
		}
		err := MetaChecks([]Doc{{Value: map[string]any{"type": "string"}}}, ext, func(c Check) error {
			calls++
			if c.Meta != sets[0].meta {
				t.Error("meta-check did not reuse the shared compiled meta-schema")
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if calls != 5 { // the callback assertion above must not be vacuous
		t.Fatalf("meta-check callback ran %d times, want 5", calls)
	}
	// A caller resource added to the per-call index and compiled against never
	// reaches the shared index.
	set, _ := sharedBuiltin()
	before := len(set.idx.Resources)
	const extra = "https://example.invalid/extra"
	ext := &registry.Index{Resources: map[string]*registry.Resource{}}
	if err := ext.AddBuiltins(); err != nil {
		t.Fatal(err)
	}
	if err := ext.AddExternal(map[string]any{"type": "integer"}, extra, "test"); err != nil {
		t.Fatal(err)
	}
	if _, ok := ext.Resources[extra]; !ok {
		t.Fatal("caller resource missing from the per-call index")
	}
	if _, _, err := CompileReached(map[string]any{"$ref": extra}, StdinBaseURI, ext); err != nil {
		t.Fatal(err)
	}
	if _, ok := set.idx.Resources[extra]; ok || len(set.idx.Resources) != before {
		t.Fatal("shared built-in index was written")
	}
	if n := builtinBuilds.Load(); n != 1 {
		t.Fatalf("buildBuiltin ran %d times, want 1", n)
	}
}
