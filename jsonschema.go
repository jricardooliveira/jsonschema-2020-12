// Package jsonschema validates JSON documents against JSON Schema 2020-12
// schemas. Compile once, then call Validate any number of times.
package jsonschema

import (
	"errors"
	"fmt"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/eval"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
	_ "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords/all"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/registry"
)

// Schema is a compiled schema, immutable and safe for concurrent use.
type Schema struct {
	root   *compile.Node
	lim    jsonvalue.NumberLimits
	max    int // output-unit cap
	maxAnn int // annotation cap
	// maxReport is the rendered-report byte cap.
	maxReport int
	maxInput  int // per-input byte cap
	el        eval.Limits
}

// Result is the outcome of one Validate call.
type Result struct {
	valid       bool
	units       []keywords.Unit
	annotations *eval.Lazy // sorted on first use
	maxReport   int
	// read-only inputs to rebuild the evaluation tree in Output
	root        *compile.Node
	inst        any
	max, maxAnn int
	el          eval.Limits
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
	s, err := compileChecked(schema, opts, reg)
	err = asLimit(err)
	var ise *InvalidSchemaError
	if errors.As(err, &ise) && ise.Source == "" {
		ise.Source = opts.SchemaPath // name the main schema file when known
	}
	return s, err
}

func compileChecked(schema []byte, opts Options, reg Registry) (*Schema, error) {
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
	rl, err := opts.resolveLimits()
	if err != nil {
		return nil, err
	}
	base, err := opts.baseURI()
	if err != nil {
		return nil, fmt.Errorf("options: %w", err)
	}
	lim.MaxDepth = rl.depth // one depth for the schema, registry files and documents
	if reg.MaxFileBytes == 0 {
		reg.MaxFileBytes = rl.input // each input shares one cap
	}
	if err := checkInput(schema, rl.input); err != nil {
		return nil, err
	}
	v, err := jsonvalue.DecodeLimited(schema, lim)
	if err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}
	ext, err := reg.load(lim)
	if err != nil {
		return nil, fmt.Errorf("registry: %w", err)
	}
	el := eval.Limits{MaxUnits: max, MaxAnnotations: maxAnn, MaxSteps: rl.steps, MaxRefChain: rl.refChain, MaxDepth: rl.depth}
	n, docs, err := compile.CompileReachedFormat(v, base, ext, rl.pattern, opts.FormatAssertion)
	if err != nil {
		// A keyword rejecting its own value is usually a meta-schema failure
		// of the main schema or of a supplied one reached before the failure:
		// report that as InvalidSchemaError naming the failing document. A
		// limit or '$schema' error from the check is reported as such.
		if merr := metaCheck(docs, ext, el); merr != nil {
			return nil, merr
		}
		return nil, fmt.Errorf("schema: %w", err)
	}
	if err := metaCheck(docs, ext, el); err != nil {
		return nil, err
	}
	n.NoUnevaluated = !compile.NeedsEvaluated(docs, ext)
	return &Schema{root: n, lim: lim, max: max, maxAnn: maxAnn, maxReport: maxRep, maxInput: rl.input, el: el}, nil
}

// metaCheck validates the main schema and every reached supplied schema,
// and each custom meta-schema on their chains, against their meta-schema
// (core 8.1.1), under the caller's configured caps (each run gets a fresh
// budget). Failing is *InvalidSchemaError naming the failing document; limit
// errors and any internal error also fail (fail closed). Format is never
// asserted here, whatever the options say.
func metaCheck(docs []compile.Doc, ext *registry.Index, el eval.Limits) error {
	err := compile.MetaChecks(docs, ext, func(c compile.Check) error {
		ev, err := eval.EvaluateNoAnnotations(c.Meta, c.Doc.Value, el)
		if err != nil {
			return err
		}
		if ev.Valid {
			return nil
		}
		var deepest keywords.Unit
		for _, u := range ev.Units {
			if len(u.KeywordLocation) >= len(deepest.KeywordLocation) {
				deepest = u
			}
		}
		return newInvalidSchemaError(c.Doc.Source, deepest.InstanceLocation, deepest.KeywordLocation)
	})
	var ise *InvalidSchemaError
	if err != nil && !errors.As(err, &ise) {
		return fmt.Errorf("schema: %w", err)
	}
	return err
}

// Validate checks a JSON document. An error means it could not be checked.
func (s *Schema) Validate(document []byte) (Result, error) {
	if err := checkInput(document, s.maxInput); err != nil {
		return Result{}, err
	}
	v, err := jsonvalue.DecodeLimited(document, s.lim)
	if err != nil {
		return Result{}, asLimit(fmt.Errorf("document: %w", err))
	}
	ev, err := eval.EvaluateValidate(s.root, v, s.el)
	if err != nil {
		return Result{}, asLimit(fmt.Errorf("document: %w", err))
	}
	return Result{valid: ev.Valid, units: ev.Units, annotations: eval.NewLazy(ev.Annotations), maxReport: s.maxReport, root: s.root, inst: v, max: s.max, maxAnn: s.maxAnn, el: s.el}, nil
}
