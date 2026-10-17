package eval

import "github.com/neoface/jsonschema/internal/keywords"

// discardsAnnotations reports whether the annotations recorded while
// evaluating keyword name can never be shown. 'not' drops its subschema's
// annotations whatever the outcome (core 10.2.1.4, 7.7.1). Never in a tree
// evaluation: detailed and verbose show the tree's annotations.
func discardsAnnotations(ctx *keywords.EvalContext, name string) bool {
	return name == "not" && ctx.Tree == nil
}

// enterSink starts discard mode and returns the function that restores the
// previous mode. Discarded annotations still count against the annotation cap.
func enterSink(ctx *keywords.EvalContext) (leave func()) {
	prev := ctx.SetDiscard(true)
	return func() { ctx.SetDiscard(prev) }
}
