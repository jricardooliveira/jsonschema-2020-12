package jsonschema

import (
	"errors"
	"fmt"

	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/registry"
)

// RegistryFile is one supplied schema file. URI is its mapped address; empty
// means it is found only by its own $id.
type RegistryFile struct {
	URI  string
	Path string
}

// RegistryDir is a supplied folder: every .json file under it is addressed as
// BaseURI plus its relative path.
type RegistryDir struct {
	BaseURI string
	Path    string
}

// Registry lists the only extra schemas a reference may resolve to. Nothing
// else is ever read. It is copied at compile time, so later changes by the
// caller have no effect.
type Registry struct {
	Files []RegistryFile
	Dirs  []RegistryDir
	// MaxFileBytes caps each supplied file; 0 means registry.DefaultMaxFileBytes.
	MaxFileBytes int
	// MaxTotalBytes caps all supplied files together; 0 means
	// registry.DefaultMaxRegistryBytes.
	MaxTotalBytes int
}

// Default registry caps, applied when the Registry fields are 0.
const (
	DefaultMaxRegistryFileBytes  = registry.DefaultMaxFileBytes
	DefaultMaxRegistryTotalBytes = registry.DefaultMaxRegistryBytes
)

// CompileWithRegistry is CompileWithOptions that also resolves references to
// the schemas listed in reg. The registry is fixed once it returns.
func CompileWithRegistry(schema []byte, opts Options, reg Registry) (*Schema, error) {
	return compileWith(schema, opts, reg)
}

// load indexes the embedded meta-schemas, then reads the listed sources and
// indexes them; the result is read-only. Built-ins go first, so a supplied
// file at a built-in address is accepted only when equal to it and the
// embedded copy stays indexed.
func (r Registry) load(lim jsonvalue.NumberLimits) (*registry.Index, error) {
	if r.MaxFileBytes < 0 || r.MaxTotalBytes < 0 {
		return nil, errors.New("options: registry limits must not be negative")
	}
	idx, err := registry.Build(true, "https://registry.invalid/unused")
	if err != nil {
		return nil, err
	}
	if err := idx.AddBuiltins(); err != nil {
		return nil, err
	}
	ld := &registry.Loader{PerFile: r.MaxFileBytes, Total: r.MaxTotalBytes, Numbers: lim}
	for _, f := range append([]RegistryFile(nil), r.Files...) {
		l, err := ld.File(f.Path, f.URI)
		if err != nil {
			return nil, err
		}
		if err := idx.AddExternal(l.Doc, l.Address, l.Path); err != nil {
			return nil, err
		}
	}
	for _, d := range append([]RegistryDir(nil), r.Dirs...) {
		ls, err := ld.Dir(d.Path, d.BaseURI)
		if err != nil {
			return nil, err
		}
		for _, l := range ls {
			if err := idx.AddExternal(l.Doc, l.Address, l.Path); err != nil {
				return nil, fmt.Errorf("%w", err)
			}
		}
	}
	return idx, nil
}
