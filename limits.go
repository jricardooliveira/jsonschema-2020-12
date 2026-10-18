package jsonschema

import (
	"errors"
	"fmt"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/eval"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords/validation"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/limits"
)

// Limit names which limit was exceeded.
type Limit string

// The closed set of limits a LimitError can name.
const (
	LimitDepth          Limit = limits.Depth
	LimitInputSize      Limit = limits.InputSize
	LimitRegistrySize   Limit = limits.RegistrySize
	LimitRefChain       Limit = limits.RefChain
	LimitSteps          Limit = limits.Steps
	LimitNumberDigits   Limit = limits.NumberDigits
	LimitNumberExponent Limit = limits.NumberExponent
	LimitPatternSize    Limit = limits.PatternSize
	LimitOutputUnits    Limit = limits.OutputUnits
	LimitAnnotations    Limit = limits.Annotations
	LimitReportBytes    Limit = limits.ReportBytes
)

// Defaults, applied when the matching Options field is 0. Limits can be
// raised or lowered, never switched off. A very large depth can exhaust the
// goroutine stack.
const (
	// DefaultMaxDepth is the deepest array/object nesting of a schema or document.
	DefaultMaxDepth = jsonvalue.MaxDepth
	// DefaultMaxInputBytes is the largest single input: a schema, a registry
	// file or a document (64 MiB).
	DefaultMaxInputBytes = 64 << 20
	// DefaultMaxRefChain is how many $ref/$dynamicRef jumps one evaluation
	// path may make (twice the depth default).
	DefaultMaxRefChain = 2 * DefaultMaxDepth
	// DefaultMaxSteps is the work-step budget of one document (100,000,000).
	DefaultMaxSteps = eval.DefaultMaxEvalSteps
	// DefaultMaxPatternSize is the longest accepted regular expression, in bytes.
	DefaultMaxPatternSize = validation.DefaultMaxPatternSize
)

// LimitError reports that a built-in limit was exceeded: the input could not
// be checked. It names the limit and the value that was set, never content.
type LimitError struct {
	Limit Limit
	Value int
	inner error
}

func (e *LimitError) Error() string {
	return fmt.Sprintf("limit exceeded: %s %d", e.Limit, e.Value)
}

// Unwrap keeps the internal error reachable for errors.As.
func (e *LimitError) Unwrap() error { return e.inner }

var closedSet = map[string]bool{
	limits.Depth: true, limits.InputSize: true, limits.RegistrySize: true, limits.RefChain: true,
	limits.Steps: true, limits.NumberDigits: true, limits.NumberExponent: true, limits.PatternSize: true,
	limits.OutputUnits: true, limits.Annotations: true, limits.ReportBytes: true,
}

// asLimit turns any internal limit error in err's chain into a top-level
// *LimitError; other errors pass through unchanged.
func asLimit(err error) error {
	var le *jsonvalue.LimitError
	// ponytail: registry reads still use their old names and file-naming
	// messages until the registry task renames them into the closed set.
	if err != nil && errors.As(err, &le) && closedSet[le.Limit] {
		return &LimitError{Limit: Limit(le.Limit), Value: le.Max, inner: err} // keep the whole chain, e.g. a *eval.LoopError
	}
	return err
}

// resolvedLimits are the Options limits after defaults, fixed at compile time.
type resolvedLimits struct {
	depth, input, refChain, steps, pattern int
}

func (o Options) resolveLimits() (r resolvedLimits, err error) {
	for _, f := range []struct {
		name   string
		v, def int
		into   *int
	}{
		{"MaxDepth", o.MaxDepth, DefaultMaxDepth, &r.depth},
		{"MaxInputBytes", o.MaxInputBytes, DefaultMaxInputBytes, &r.input},
		{"MaxRefChain", o.MaxRefChain, DefaultMaxRefChain, &r.refChain},
		{"MaxSteps", o.MaxSteps, DefaultMaxSteps, &r.steps},
		{"MaxPatternSize", o.MaxPatternSize, DefaultMaxPatternSize, &r.pattern},
	} {
		if *f.into, err = limits.Resolve(f.name, f.v, f.def); err != nil {
			return resolvedLimits{}, err
		}
	}
	return r, nil
}

// checkInput fails with a LimitError when data is over the input cap.
func checkInput(data []byte, max int) error {
	if len(data) > max {
		return &LimitError{Limit: LimitInputSize, Value: max, inner: &jsonvalue.LimitError{Limit: limits.InputSize, Max: max}}
	}
	return nil
}
