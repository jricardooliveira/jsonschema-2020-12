// Package output is the one registration point for report shapes. A shape
// file calls Register or RegisterReport from init(); callers find shapes only
// through Render and RenderReport.
package output

import (
	"bytes"
	"fmt"

	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
)

// DefaultMaxReportBytes caps the size of one rendered report.
const DefaultMaxReportBytes = 256 << 20

// Renderer turns a validation outcome and its failure units into JSON.
type Renderer func(valid bool, units []keywords.Unit) ([]byte, error)

// Report is everything a shape may render. Annotations is empty unless Valid.
type Report struct {
	Valid       bool
	Units       []keywords.Unit
	Annotations []keywords.Annotation
	// Tree is the evaluation tree; set only for shapes registered with
	// RegisterTreeReport. Nil means "render a minimal root unit".
	Tree *keywords.TreeNode
}

// needsTree names the shapes that want Report.Tree. Written by init() only.
var needsTree = map[string]bool{}

// RegisterTreeReport adds a shape that renders from the evaluation tree.
func RegisterTreeReport(name string, r ReportRenderer) {
	RegisterReport(name, r)
	needsTree[name] = true
}

// NeedsTree reports whether the named shape renders from Report.Tree.
func NeedsTree(name string) bool { return needsTree[name] }

// Budget is a byte buffer that refuses to grow past its cap.
type Budget struct {
	buf bytes.Buffer
	max int
}

// Write appends p, or returns a "report bytes" LimitError (writing nothing)
// when the total would exceed the cap.
func (b *Budget) Write(p []byte) (int, error) {
	if b.buf.Len()+len(p) > b.max {
		return 0, &jsonvalue.LimitError{Limit: "report bytes", Max: b.max}
	}
	return b.buf.Write(p)
}

// ReportRenderer writes a report into the budget.
type ReportRenderer func(r Report, w *Budget) error

// registry is written only by init() functions and read-only afterwards.
var registry = map[string]ReportRenderer{}

// RegisterReport adds a shape. It panics on an empty name, nil renderer or duplicate.
func RegisterReport(name string, r ReportRenderer) {
	if name == "" || r == nil {
		panic("output: incomplete registration")
	}
	if _, dup := registry[name]; dup {
		panic("output: duplicate registration of " + name)
	}
	registry[name] = r
}

// Register adds an old-style shape; its bytes still go through the budget.
func Register(name string, r Renderer) {
	if r == nil {
		panic("output: incomplete registration")
	}
	RegisterReport(name, func(rep Report, w *Budget) error {
		b, err := r(rep.Valid, rep.Units)
		if err != nil {
			return err
		}
		_, err = w.Write(b)
		return err
	})
}

// RenderReport renders the named shape within maxBytes. A cap of 0 or less
// means DefaultMaxReportBytes, never unlimited. No partial report is returned.
func RenderReport(name string, r Report, maxBytes int) ([]byte, error) {
	fn, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown output format %q", name)
	}
	if maxBytes <= 0 {
		maxBytes = DefaultMaxReportBytes
	}
	w := &Budget{max: maxBytes}
	if err := fn(r, w); err != nil {
		return nil, err
	}
	return w.buf.Bytes(), nil
}

// Render renders the named shape under the default byte cap.
func Render(name string, valid bool, units []keywords.Unit) ([]byte, error) {
	return RenderReport(name, Report{Valid: valid, Units: units}, DefaultMaxReportBytes)
}
