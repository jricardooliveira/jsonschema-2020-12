package output

// The 'verbose' shape (core 12.4.3) renders the whole evaluation tree: every
// evaluated subschema node, passing or failing, with its errors and kept
// annotations.
//
// Partly produced:
//   - A passing keyword with no subschema and no annotation (for example a
//     passing minLength) has no tree node, so it gets no unit of its own;
//     recording passing leaf units on every Validate call is not done.
//   - absoluteKeywordLocation appears only where the tree node or unit
//     already carries one.
//   - Annotation values come from the schema; matched property names or
//     indices are treated like instanceLocation. Instance values never appear.
func init() {
	RegisterTreeReport("verbose", func(r Report, w *Budget) error { return renderTree(r, w, true) })
}
