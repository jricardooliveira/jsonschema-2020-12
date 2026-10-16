package eval

import (
	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/keywords"
)

// attachItems gives a root context its per-call record; child contexts
// inherit the parent's (keywords.EvalContext.EvalChild).
func attachItems(ctx *keywords.EvalContext) {
	if ctx.Items == nil {
		ctx.Items = NewEvaluatedItems()
	}
}

// EvaluateAnnotated is Evaluate that also returns the item-annotation record.
// A failed root schema keeps no annotations (core 7.7.1).
func EvaluateAnnotated(n *compile.Node, instance any) (bool, *EvaluatedItems, error) {
	ctx := keywords.NewEvalContext(DefaultMaxOutputUnits, evalNode)
	rec := NewEvaluatedItems()
	ctx.Items = rec
	ok, err := evalNode(ctx, n, instance)
	if !ok {
		rec.Rollback(0)
	}
	return ok, rec, err
}

// ItemAnnotation is what prefixItems, items or contains recorded for one
// array (2020-12 core 10.3.1).
type ItemAnnotation struct {
	Keyword string // "prefixItems", "items" or "contains"
	// All is the annotation value true: prefixItems applied to every element,
	// items applied, or contains matched every element.
	All bool
	// Largest is prefixItems' largest applied index (when All is false).
	Largest int
	// Indices are contains' matching indices (when All is false).
	Indices []int
}

type itemKey struct{ instance, keyword string }

// EvaluatedItems is the per-evaluation record of item annotations. It lives
// in one evaluation's context, never in a compiled schema, and is used by one
// goroutine. Entries are append-only so a failed branch can be rolled back
// with Mark and Rollback.
type EvaluatedItems struct {
	entries []itemEntry
	// steps counts evaluation steps for this call (see Step); Rollback never
	// resets it.
	steps, maxSteps int
}

type itemEntry struct {
	key itemKey
	ann ItemAnnotation
	// props marks a property-name entry (see RecordProperties); names holds them.
	props bool
	names []string
}

// NewEvaluatedItems returns an empty record.
func NewEvaluatedItems() *EvaluatedItems { return newEvaluatedItems(DefaultMaxEvalSteps) }

func newEvaluatedItems(maxSteps int) *EvaluatedItems { return &EvaluatedItems{maxSteps: maxSteps} }

// Mark returns a position to pass to Rollback.
func (r *EvaluatedItems) Mark() int { return len(r.entries) }

// Rollback drops annotations recorded after mark (a failed branch must not
// leak them).
func (r *EvaluatedItems) Rollback(mark int) {
	if mark >= 0 && mark < len(r.entries) {
		r.entries = r.entries[:mark]
	}
}

func (r *EvaluatedItems) add(inst, kw string, a ItemAnnotation) {
	a.Keyword = kw
	r.entries = append(r.entries, itemEntry{key: itemKey{inst, kw}, ann: a})
}

// RecordPrefixItems records prefixItems applied to elements 0..largest; all
// is true when it applied to every element.
func (r *EvaluatedItems) RecordPrefixItems(instLoc, kwLoc string, largest int, all bool) {
	r.add(instLoc, kwLoc, ItemAnnotation{Largest: largest, All: all})
}

// RecordItems records items applied to at least one element (value true).
func (r *EvaluatedItems) RecordItems(instLoc, kwLoc string) {
	r.add(instLoc, kwLoc, ItemAnnotation{All: true})
}

// RecordContains records contains' matching indices, or true when all matched.
// The slice is copied.
func (r *EvaluatedItems) RecordContains(instLoc, kwLoc string, indices []int, all bool) {
	a := ItemAnnotation{All: all}
	if !all {
		a.Indices = append([]int(nil), indices...)
	}
	r.add(instLoc, kwLoc, a)
}

// At returns the annotations recorded for the array at instLoc, in the order
// they were recorded. kwLoc of each is the keyword location it was recorded at.
func (r *EvaluatedItems) At(instLoc string) map[string]ItemAnnotation {
	out := map[string]ItemAnnotation{}
	for _, e := range r.entries {
		if !e.props && e.key.instance == instLoc {
			out[e.key.keyword] = e.ann
		}
	}
	return out
}
