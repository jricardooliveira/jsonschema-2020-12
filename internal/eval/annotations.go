package eval

import (
	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/keywords"
)

// DefaultMaxAnnotations is the default cap on annotations one evaluation may
// record (dropped ones count too). Going over is a *jsonvalue.LimitError
// named "annotations".
const DefaultMaxAnnotations = keywords.DefaultMaxAnnotations

// Annotation and Tree are the per-call annotation record and the
// per-subschema evaluation tree.
type (
	Annotation = keywords.Annotation
	Tree       = keywords.TreeNode
)

// Evaluation is everything one evaluation produced. Annotations is sorted
// (see SortAnnotations) and empty unless Valid. Tree (EvaluateTree only)
// holds every evaluated subschema, including passing parts of a failing
// evaluation; annotations dropped by a failing subschema, and by everything
// below it, appear nowhere in it. A failing root keeps none of its own.
type Evaluation struct {
	Valid       bool
	Units       []keywords.Unit
	Annotations []Annotation
	// Pending is set instead of Annotations by a lazy evaluation
	// (EvaluateLazy, EvaluateValidate): the unflattened chunks, flattened
	// and sorted on the first Sorted call. Nil unless Valid.
	Pending *Pending
	Tree    *Tree
}

// EvaluateFull evaluates n with explicit output-unit and annotation caps. It
// builds no evaluation tree; annotations are always collected.
func EvaluateFull(n *compile.Node, instance any, maxUnits, maxAnnotations int) (Evaluation, error) {
	return evaluate(n, instance, maxUnits, maxAnnotations, nil, nil)
}

// EvaluateTree is EvaluateFull that also returns the per-subschema
// evaluation tree (for the detailed and verbose shapes).
func EvaluateTree(n *compile.Node, instance any, maxUnits, maxAnnotations int) (Evaluation, error) {
	return evaluate(n, instance, maxUnits, maxAnnotations, &Tree{}, nil)
}

// evaluate runs one call; state, when set, is the per-call record (with its
// caps) to use instead of a default one.
func evaluate(n *compile.Node, instance any, maxUnits, maxAnnotations int, tree *Tree, state any) (Evaluation, error) {
	return evaluateOpt(n, instance, maxUnits, maxAnnotations, tree, state, false)
}

// evaluateOpt is evaluate; lazy leaves the annotations unflattened and
// unsorted in Pending.
func evaluateOpt(n *compile.Node, instance any, maxUnits, maxAnnotations int, tree *Tree, state any, lazy bool) (Evaluation, error) {
	ctx := keywords.NewEvalContext(maxUnits, evalNode)
	ctx.Items = state
	ctx.SetMaxAnnotations(maxAnnotations)
	ctx.Tree = tree
	ok, err := evalNode(ctx, n, instance)
	if err != nil {
		return Evaluation{}, err
	}
	ev := Evaluation{Valid: ok, Units: ctx.Units(), Tree: tree}
	switch {
	case ok && lazy:
		ev.Pending = &Pending{chunks: ctx.AnnotationChunks()}
	case ok:
		ev.Annotations = sortInPlace(ctx.Annotations()) // already a fresh copy
	}
	if tree != nil {
		tree.Valid, tree.Errors = ok, ctx.Units()
		if !ok {
			tree.Annotations = nil
		}
	}
	return ev, nil
}

// SortAnnotations returns a sorted copy: by instance location, then keyword
// location (both compared as decoded JSON Pointer tokens in byte order, a
// shorter prefix first), then absolute keyword location, keyword name and
// evaluation sequence.
func SortAnnotations(in []Annotation) []Annotation {
	return sortAnnotations(in)
}
