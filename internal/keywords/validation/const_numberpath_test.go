package validation_test

// Number paths for const: validation spec section 6.1.3, using equalCases.

import "testing"

func TestConstNumberPaths(t *testing.T) {
	for _, c := range equalCases {
		if got := npValid(t, `{"const":`+c.val+`}`, c.doc); got != c.want {
			t.Errorf("const %s doc %s: got %v want %v", c.val, c.doc, got, c.want)
		}
	}
}
