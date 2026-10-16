package eval

import (
	"slices"
	"strings"

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
// (see SortAnnotations) and empty unless Valid. Tree holds every evaluated
// subschema, including passing parts of a failing evaluation; annotations
// dropped by a failing subschema appear nowhere in it.
type Evaluation struct {
	Valid       bool
	Units       []keywords.Unit
	Annotations []Annotation
	Tree        *Tree
}

// EvaluateFull evaluates n with explicit output-unit and annotation caps.
func EvaluateFull(n *compile.Node, instance any, maxUnits, maxAnnotations int) (Evaluation, error) {
	ctx := keywords.NewEvalContext(maxUnits, evalNode)
	ctx.SetMaxAnnotations(maxAnnotations)
	ctx.Tree = &Tree{}
	ok, err := evalNode(ctx, n, instance)
	if err != nil {
		return Evaluation{}, err
	}
	ctx.Tree.Valid, ctx.Tree.Errors = ok, ctx.Units()
	ev := Evaluation{Valid: ok, Units: ctx.Units(), Tree: ctx.Tree}
	if ok {
		ev.Annotations = SortAnnotations(ctx.Annotations())
	} else {
		ctx.Tree.Annotations = nil
	}
	return ev, nil
}

// SortAnnotations returns a sorted copy: by instance location, then keyword
// location (both compared as decoded JSON Pointer tokens in byte order, a
// shorter prefix first), then absolute keyword location, keyword name and
// evaluation sequence.
func SortAnnotations(in []Annotation) []Annotation {
	out := slices.Clone(in)
	slices.SortStableFunc(out, func(a, b Annotation) int {
		if c := cmpPointer(a.InstanceLocation, b.InstanceLocation); c != 0 {
			return c
		}
		if c := cmpPointer(a.KeywordLocation, b.KeywordLocation); c != 0 {
			return c
		}
		if c := strings.Compare(a.AbsoluteKeywordLocation, b.AbsoluteKeywordLocation); c != 0 {
			return c
		}
		if c := strings.Compare(a.Keyword, b.Keyword); c != 0 {
			return c
		}
		return a.Seq - b.Seq
	})
	return out
}

var unescape = strings.NewReplacer("~1", "/", "~0", "~")

func tokens(p string) []string {
	if p == "" {
		return nil
	}
	parts := strings.Split(p[1:], "/")
	for i, t := range parts {
		parts[i] = unescape.Replace(t)
	}
	return parts
}

func cmpPointer(a, b string) int { return slices.Compare(tokens(a), tokens(b)) }
