package validation_test

// Number paths for exclusiveMaximum (validation spec section 6.2.3), using boundCases.

import "testing"

func TestExclusiveMaximumNumberPaths(t *testing.T) {
	runBoundCases(t, "exclusiveMaximum", func(c int) bool { return c < 0 })
}
