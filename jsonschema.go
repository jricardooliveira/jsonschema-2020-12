// Package jsonschema validates JSON documents against JSON Schema 2020-12
// schemas. Compile once, then call Validate any number of times.
package jsonschema

import (
	"fmt"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/eval"
	"github.com/neoface/jsonschema/internal/jsonvalue"
	_ "github.com/neoface/jsonschema/internal/keywords/all"
)

// Schema is a compiled schema, immutable and safe for concurrent use.
type Schema struct {
	root *compile.Node
	lim  jsonvalue.NumberLimits
}

// Result is the outcome of one Validate call.
type Result struct{ valid bool }

// Valid reports whether the document satisfied the schema.
func (r Result) Valid() bool { return r.valid }

// Compile parses and compiles a schema. Broken JSON, limit errors and
// schemas that are not a boolean or object are errors.
func Compile(schema []byte) (*Schema, error) { return CompileWithOptions(schema, Options{}) }

// CompileWithOptions is Compile with explicit Options.
func CompileWithOptions(schema []byte, opts Options) (*Schema, error) {
	lim, err := opts.limits()
	if err != nil {
		return nil, err
	}
	v, err := jsonvalue.DecodeLimited(schema, lim)
	if err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}
	n, err := compile.Compile(v)
	if err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}
	return &Schema{root: n, lim: lim}, nil
}

// Validate checks a JSON document. An error means it could not be checked.
func (s *Schema) Validate(document []byte) (Result, error) {
	v, err := jsonvalue.DecodeLimited(document, s.lim)
	if err != nil {
		return Result{}, fmt.Errorf("document: %w", err)
	}
	ok, err := eval.Evaluate(s.root, v)
	if err != nil {
		return Result{}, fmt.Errorf("document: %w", err)
	}
	return Result{valid: ok}, nil
}
