package validation_test

// Number paths for const: validation spec section 6.1.3, using equalCases.

import "testing"

func TestConstNumberPaths(t *testing.T) {
	for _, c := range append(equalCases, equalCapCases()...) {
		if got := npValid(t, `{"const":`+c.val+`}`, c.doc); got != c.want {
			t.Errorf("const %.30s doc %.30s: got %v want %v", c.val, c.doc, got, c.want)
		}
	}
	npLimit(t, `{"const":1}`)
}
