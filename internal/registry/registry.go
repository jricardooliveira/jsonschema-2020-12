package registry

import (
	"fmt"
	"sort"
	"sync"

	"github.com/neoface/jsonschema/internal/jsonvalue"
)

// BuiltinSource is the source label of the embedded meta-schemas in errors.
const BuiltinSource = "built-in meta-schema"

// BuildBuiltinIndex builds, from scratch, an index holding only the embedded
// meta-schemas under their $id. It is not cached: callers that share the
// result must treat it as read-only.
func BuildBuiltinIndex() (*Index, error) {
	builtinBuilds.Add(1)
	idx, err := Build(true, "https://registry.invalid/unused")
	if err != nil {
		return nil, err
	}
	docs, err := Builtins()
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(docs))
	for id := range docs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if err := idx.AddExternal(docs[id], id, BuiltinSource); err != nil {
			return nil, err
		}
	}
	return idx, nil
}

var (
	sharedOnce sync.Once
	sharedIdx  *Index
	sharedErr  error
)

// SharedBuiltinIndex returns the process-wide built-in index, built once on
// first use. It is immutable after the build: callers must never write to it
// or to its resources, and may share it across goroutines.
func SharedBuiltinIndex() (*Index, error) {
	sharedOnce.Do(func() { sharedIdx, sharedErr = BuildBuiltinIndex() })
	return sharedIdx, sharedErr
}

// AddBuiltins indexes the embedded meta-schemas under their $id. Call it
// before adding supplied documents: AddExternal then accepts a supplied file at
// a built-in address only when it is deep-equal to the embedded one (the
// embedded copy stays indexed) and refuses a different one with an error that
// names the address and the file, quoting no content. The built-in resources
// are built once per process and shared read-only; idx itself is only written
// here, never the shared resources.
func (idx *Index) AddBuiltins() error {
	if _, err := SharedBuiltinIndex(); err != nil {
		return err
	}
	keys := make([]string, 0, len(sharedIdx.Resources))
	for k, r := range sharedIdx.Resources {
		if r.Source == BuiltinSource {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		r := sharedIdx.Resources[k]
		if prev, dup := idx.Resources[k]; dup {
			if prev != r && !jsonvalue.Equal(prev.value(), r.value()) {
				return fmt.Errorf("duplicate address %q claimed by %s and %s", k, prev.from(), r.from())
			}
			continue
		}
		idx.Resources[k] = r
	}
	return nil
}
