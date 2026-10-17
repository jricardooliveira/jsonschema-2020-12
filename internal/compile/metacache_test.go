package compile

import (
	"sync"
	"testing"

	"github.com/neoface/jsonschema/internal/registry"
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
	for i := 0; i < 5; i++ {
		ext := &registry.Index{Resources: map[string]*registry.Resource{}}
		if err := ext.AddBuiltins(); err != nil {
			t.Fatal(err)
		}
		err := MetaChecks([]Doc{{Value: map[string]any{"type": "string"}}}, ext, func(c Check) error {
			if c.Meta != sets[0].meta {
				t.Error("meta-check did not reuse the shared compiled meta-schema")
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if n := builtinBuilds.Load(); n != 1 {
		t.Fatalf("buildBuiltin ran %d times, want 1", n)
	}
}
