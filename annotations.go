package jsonschema

import "github.com/neoface/jsonschema/internal/eval"

// DefaultMaxAnnotations is the default cap on annotations one Validate may
// record (dropped ones count too). Exceeding it is a LimitError.
const DefaultMaxAnnotations = eval.DefaultMaxAnnotations

// Annotation is one collected annotation. Value comes from the schema, except
// that properties, patternProperties and additionalProperties record the
// evaluated property NAMES, which already appear in the instance location.
// An instance value is never an annotation value. Treat Value as read only.
type Annotation struct {
	InstanceLocation        string
	KeywordLocation         string
	AbsoluteKeywordLocation string // empty when not known
	Keyword                 string
	Value                   any
}

// Annotations returns the annotations of a passing result, ordered by
// instance location then keyword location; a failing result has none.
func (r Result) Annotations() []Annotation {
	out := make([]Annotation, len(r.annotations))
	for i, a := range r.annotations {
		out[i] = Annotation{a.InstanceLocation, a.KeywordLocation, a.AbsoluteKeywordLocation, a.Keyword, a.Value}
	}
	return out
}
