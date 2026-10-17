package compile

import (
	"errors"

	"github.com/neoface/jsonschema/internal/registry"
	"github.com/neoface/jsonschema/internal/uri"
)

// Draft2020 is the only supported meta-schema root (core 8.1.1).
const Draft2020 = "https://json-schema.org/draft/2020-12/schema"

// maxMetaChain bounds the custom meta-schema chain so hostile input ends.
const maxMetaChain = 32

// Errors from ResolveMetaSchema. Their text never echoes the declared '$schema'.
var (
	// ErrUnsupportedSchema: '$schema' names a version or meta-schema that is not
	// supported (not 2020-12 and not a supplied meta-schema).
	ErrUnsupportedSchema = errors.New("'$schema': the declared meta-schema version is not supported")
	// ErrMetaChain: a custom meta-schema chain loops, is too long, or cannot be read.
	ErrMetaChain = errors.New("'$schema': the meta-schema chain does not end in the 2020-12 meta-schema")
)

// declared returns the normalized '$schema' of a schema value (empty fragment
// dropped), or "" when absent or not a string (the meta-check reports that).
func declared(v any) (string, error) {
	m, _ := v.(map[string]any)
	s, ok := m["$schema"].(string)
	if !ok {
		return "", nil
	}
	n, err := uri.Normalize(s)
	if err != nil {
		return "", ErrUnsupportedSchema
	}
	if len(n) > 0 && n[len(n)-1] == '#' {
		n = n[:len(n)-1]
	}
	return n, nil
}

// ResolveMetaSchema picks the meta-schema URI a schema is checked against
// (core 8.1.1): Draft2020 when '$schema' is absent or names 2020-12, or the
// URI of a custom meta-schema found in idx whose own chain ends in 2020-12.
// Cycles and chains longer than maxMetaChain are ErrMetaChain.
func ResolveMetaSchema(idx *registry.Index, schema any) (string, error) {
	cur, err := declared(schema)
	if err != nil {
		return "", err
	}
	if cur == "" {
		return Draft2020, nil
	}
	first := cur
	seen := map[string]bool{}
	for {
		if cur == Draft2020 {
			return first, nil
		}
		if seen[cur] || len(seen) >= maxMetaChain {
			return "", ErrMetaChain
		}
		seen[cur] = true
		var r *registry.Resource
		if idx != nil {
			r = idx.Resources[cur]
		}
		if r == nil {
			if len(seen) == 1 {
				return "", ErrUnsupportedSchema
			}
			return "", ErrMetaChain
		}
		if r.Pointer != "" {
			return "", ErrMetaChain
		}
		if cur, err = declared(r.Doc); err != nil {
			return "", ErrMetaChain
		}
		if cur == "" {
			return first, nil // no '$schema' means 2020-12
		}
	}
}
