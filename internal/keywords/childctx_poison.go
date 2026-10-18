package keywords

import "testing"

// poison wipes a released child context in test binaries only, so a keyword
// that kept it past EvalChild fails loudly (nil dereference in Report and
// EvalChild) instead of silently seeing a sibling's state. In production
// testing.Testing() is false and this does nothing.
func poison(c *EvalContext) {
	if testing.Testing() {
		*c = EvalContext{InstanceLocation: "<poisoned child context>"}
	}
}
