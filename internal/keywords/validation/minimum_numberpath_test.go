package validation_test

// Number paths for minimum (validation spec section 6.2.4), using boundCases.

import "testing"

func TestMinimumNumberPaths(t *testing.T) {
	runBoundCases(t, "minimum", func(c int) bool { return c >= 0 })
}
