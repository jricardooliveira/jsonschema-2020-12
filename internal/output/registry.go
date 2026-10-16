// Package output is the one registration point for report shapes. A shape
// file calls Register from init(); callers find shapes only through Render.
package output

import (
	"fmt"

	"github.com/neoface/jsonschema/internal/keywords"
)

// Renderer turns a validation outcome and its failure units into JSON.
type Renderer func(valid bool, units []keywords.Unit) ([]byte, error)

// registry is written only by init() functions and read-only afterwards.
var registry = map[string]Renderer{}

// Register adds a shape. It panics on an empty name, nil renderer or duplicate.
func Register(name string, r Renderer) {
	if name == "" || r == nil {
		panic("output: incomplete registration")
	}
	if _, dup := registry[name]; dup {
		panic("output: duplicate registration of " + name)
	}
	registry[name] = r
}

// Render renders the named shape. An unknown name is an error.
func Render(name string, valid bool, units []keywords.Unit) ([]byte, error) {
	r, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown output format %q", name)
	}
	return r(valid, units)
}
