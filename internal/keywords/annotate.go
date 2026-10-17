package keywords

import "github.com/neoface/jsonschema/internal/jsonvalue"

// DefaultMaxAnnotations caps how many annotations one evaluation may record,
// counting ones that are later dropped.
const DefaultMaxAnnotations = 10_000_000

// Annotation is one collected annotation. Value comes from the schema (by
// reference, never copied), except that properties, patternProperties and
// additionalProperties record evaluated property NAMES, which already appear
// in the instance location. An instance value is never an annotation value.
type Annotation struct {
	InstanceLocation        string
	KeywordLocation         string
	AbsoluteKeywordLocation string // empty when not known
	Keyword                 string
	Value                   any
	Seq                     int // evaluation order, the last sort tie-break
}

// TreeNode is the evaluation result of one subschema. Annotations holds the
// ones recorded directly by this node, and only when it is valid.
type TreeNode struct {
	KeywordLocation         string
	AbsoluteKeywordLocation string
	InstanceLocation        string
	Valid                   bool
	Errors                  []Unit
	Annotations             []Annotation
	Children                []*TreeNode
}

// dropAnnotations clears the annotations of n and its whole subtree; n may be
// nil. A failing child was already cleared when it failed, so it is skipped.
func (n *TreeNode) dropAnnotations() {
	if n == nil {
		return
	}
	n.Annotations = nil
	for _, c := range n.Children {
		if c.Valid {
			c.dropAnnotations()
		}
	}
}

// collector is the per-call annotation log, shared by a context and all of
// its children. It is append-only so a failed subschema can roll back.
type collector struct {
	list     []Annotation
	count    int
	max, seq int
	discard  bool // annotations are counted but neither recorded nor kept
}

// SetDiscard switches discard mode and returns the previous setting. In it
// Annotate still counts against the cap but records nothing. It is shared by
// every child context of the call.
func (c *EvalContext) SetDiscard(on bool) (prev bool) {
	prev, c.col.discard = c.col.discard, on
	return prev
}

// SetMaxAnnotations sets the annotation cap; it must be positive. Call it on
// the root context before evaluating.
func (c *EvalContext) SetMaxAnnotations(max int) { c.col.max = max }

// Annotate records an annotation for the keyword being evaluated, at the
// current locations. value must come from the schema. Every call is counted,
// even if the annotation is dropped later; over the cap it returns a LimitError.
func (c *EvalContext) Annotate(value any) error {
	col := c.col
	col.count++
	if col.count > col.max {
		return &jsonvalue.LimitError{Limit: "annotations", Max: col.max}
	}
	if col.discard {
		return nil
	}
	col.seq++
	a := Annotation{InstanceLocation: c.InstanceLocation, KeywordLocation: c.KeywordLoc(), Keyword: c.Keyword, Value: value, Seq: col.seq}
	col.list = append(col.list, a)
	if c.Tree != nil {
		c.Tree.Annotations = append(c.Tree.Annotations, a)
	}
	return nil
}

// AnnotateNames is Annotate for a member-name list. It does not copy: the
// caller must pass a slice it will never modify or reuse afterwards (never a
// stack buffer or a slice that is later appended to).
func (c *EvalContext) AnnotateNames(names []string) error { return c.Annotate(names) }

// Annotations returns the annotations kept so far, in evaluation order.
func (c *EvalContext) Annotations() []Annotation { return c.col.list }

// AnnotateOnly returns a keyword that always passes and annotates its own
// schema value. Registered or not, it never fails validation.
func AnnotateOnly(name, vocabulary string) Keyword {
	return Keyword{
		Name:       name,
		Vocabulary: vocabulary,
		Compile:    func(_ *CompileContext, value any) (any, error) { return value, nil },
		Evaluate: func(ctx *EvalContext, compiled any, _ any) (bool, error) {
			return true, ctx.Annotate(compiled)
		},
	}
}

// RegisterAnnotationOnly registers an annotate-only keyword.
func RegisterAnnotationOnly(name, vocabulary string) { Register(AnnotateOnly(name, vocabulary)) }
