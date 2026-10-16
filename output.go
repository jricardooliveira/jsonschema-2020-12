package jsonschema

import (
	"github.com/neoface/jsonschema/internal/eval"
	"github.com/neoface/jsonschema/internal/output"
)

// DefaultMaxOutputUnits is the default cap on failure units one Validate may
// append (discarded units count too). Exceeding it is a LimitError.
const DefaultMaxOutputUnits = eval.DefaultMaxOutputUnits

// Output renders the result as a 2020-12 output object. shape is "flag" or
// "basic"; an unknown shape is an error. The output never holds instance values.
func (r Result) Output(shape string) ([]byte, error) {
	return output.Render(shape, r.valid, r.units)
}
