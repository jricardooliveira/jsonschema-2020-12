package jsonschema

import (
	"github.com/jricardooliveira/jsonschema-2020-12/internal/eval"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/output"
)

// DefaultMaxOutputUnits is the default cap on failure units one Validate may
// append (discarded units count too). Exceeding it is a LimitError.
const DefaultMaxOutputUnits = eval.DefaultMaxOutputUnits

// Output renders the result as a 2020-12 output object. shape is "flag" or
// "basic"; an unknown shape is an error. The output never holds instance values.
func (r Result) Output(shape string) ([]byte, error) {
	rep := output.Report{Valid: r.valid, Units: r.units, Annotations: r.annotations.Sorted()}
	if output.NeedsTree(shape) && r.root != nil {
		// The tree is rebuilt per call, never cached, so a Result stays immutable.
		ev, err := eval.EvaluateWith(r.root, r.inst, r.el, true)
		if err != nil {
			return nil, asLimit(err)
		}
		rep.Tree = ev.Tree
	}
	b, err := output.RenderReport(shape, rep, r.maxReport)
	return b, asLimit(err)
}

// DefaultMaxReportBytes is the default cap on one rendered report. Exceeding
// it is a LimitError; a report is never truncated.
const DefaultMaxReportBytes = output.DefaultMaxReportBytes
