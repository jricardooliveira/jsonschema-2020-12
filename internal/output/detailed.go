package output

// The 'detailed' shape (core 12.4.2) renders the evaluation tree condensed:
// passing nodes without annotations are dropped, a node that only wraps one
// child is replaced by that child, and annotations under a failing node are
// not kept.
//
// Partly produced:
//   - A passing keyword with no subschema and no annotation (for example a
//     passing minLength) has no tree node, so it never appears; detailed
//     would drop it anyway.
//   - absoluteKeywordLocation appears only where the tree node or unit
//     already carries one.
//   - Annotation values come from the schema; matched property names or
//     indices are treated like instanceLocation. Instance values never appear.
func init() {
	RegisterTreeReport("detailed", func(r Report, w *Budget) error { return renderTree(r, w, false) })
}
