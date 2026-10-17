package compile

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/neoface/jsonschema/internal/registry"
)

// builtinSet is the compiled 2020-12 meta-schema and the index of the embedded
// meta-schemas it was compiled against. It depends on no caller option, is
// built once on first use and is never written afterwards.
type builtinSet struct {
	idx  *registry.Index
	meta *Node
}

var (
	builtinOnce  sync.Once
	builtinShare *builtinSet
	builtinErr   error
	// builtinBuilds counts buildBuiltin calls (tests read it).
	builtinBuilds atomic.Int32
)

// buildBuiltin builds the set from scratch (the first-load benchmark calls it
// directly). The index comes from the registry's shared builder, so the
// process builds it once. Meta-checks never assert format, so noFormat is on.
func buildBuiltin() (*builtinSet, error) {
	builtinBuilds.Add(1)
	idx, err := registry.SharedBuiltinIndex()
	if err != nil {
		return nil, err
	}
	_, n, err := compileDocumentOpts(map[string]any{"$ref": Draft2020}, StdinBaseURI, idx, 0, false, true)
	if err != nil {
		return nil, err
	}
	return &builtinSet{idx: idx, meta: n}, nil
}

// sharedBuiltin returns the process-wide set, building it on first use.
func sharedBuiltin() (*builtinSet, error) {
	builtinOnce.Do(func() { builtinShare, builtinErr = buildBuiltin() })
	if builtinShare == nil && builtinErr == nil { // build panicked
		return nil, errors.New("built-in meta-schemas are unavailable")
	}
	return builtinShare, builtinErr
}
