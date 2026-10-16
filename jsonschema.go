// Package jsonschema validates JSON documents against JSON Schema 2020-12
// schemas. Compile once, then call Validate any number of times.
package jsonschema

import (
	"fmt"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/eval"
	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
	_ "github.com/neoface/jsonschema/internal/keywords/all"
)

// Schema is a compiled schema, immutable and safe for concurrent use.
type Schema struct {
	root   *compile.Node
	lim    jsonvalue.NumberLimits
	max    int // output-unit cap
	maxAnn int // annotation cap
	// maxReport is the rendered-report byte cap.
	maxReport int
}

// Result is the outcome of one Validate call.
type Result struct {
	valid       bool
	units       []keywords.Unit
	annotations []eval.Annotation
	maxReport   int
}

// Valid reports whether the document satisfied the schema.
func (r Result) Valid() bool { return r.valid }

// Compile parses and compiles a schema. Broken JSON, limit errors and
// schemas that are not a boolean or object are errors.
func Compile(schema []byte) (*Schema, error) { return CompileWithOptions(schema, Options{}) }

// CompileWithOptions is Compile with explicit Options.
func CompileWithOptions(schema []byte, opts Options) (*Schema, error) {
	return compileWith(schema, opts, Registry{})
}

func compileWith(schema []byte, opts Options, reg Registry) (*Schema, error) {
	lim, err := opts.limits()
	if err != nil {
		return nil, err
	}
	max, err := opts.maxUnits()
	if err != nil {
		return nil, err
	}
	maxAnn, err := opts.maxAnnotations()
	if err != nil {
		return nil, err
	}
	maxRep, err := opts.maxReportBytes()
	if err != nil {
		return nil, err
	}
	base, err := opts.baseURI()
	if err != nil {
		return nil, fmt.Errorf("options: %w", err)
	}
	v, err := jsonvalue.DecodeLimited(schema, lim)
	if err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}
	ext, err := reg.load(lim)
	if err != nil {
		return nil, fmt.Errorf("registry: %w", err)
	}
	n, err := compile.CompileWithRegistry(v, base, ext)
	if err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}
	return &Schema{root: n, lim: lim, max: max, maxAnn: maxAnn, maxReport: maxRep}, nil
}

// Validate checks a JSON document. An error means it could not be checked.
func (s *Schema) Validate(document []byte) (Result, error) {
	v, err := jsonvalue.DecodeLimited(document, s.lim)
	if err != nil {
		return Result{}, fmt.Errorf("document: %w", err)
	}
	ev, err := eval.EvaluateFull(s.root, v, s.max, s.maxAnn)
	if err != nil {
		return Result{}, fmt.Errorf("document: %w", err)
	}
	return Result{valid: ev.Valid, units: ev.Units, annotations: ev.Annotations, maxReport: s.maxReport}, nil
}
