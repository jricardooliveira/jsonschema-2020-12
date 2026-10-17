package validation_test

// Number paths for exclusiveMinimum (validation spec section 6.2.5), using boundCases.

import "testing"

func TestExclusiveMinimumNumberPaths(t *testing.T) {
	runBoundCases(t, "exclusiveMinimum", func(c int) bool { return c > 0 })
}
