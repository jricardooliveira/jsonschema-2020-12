package validation_test

// Derived from JSON Schema validation spec section 6.2.4 (minimum).

import "testing"

func TestMinimumBounds(t *testing.T) {
	runBounds(t, []struct {
		schema, doc string
		want        bool
	}{
		{`{"minimum":1}`, `1.0`, true},
		{`{"minimum":1}`, `0.9999999999999999999`, false},
		{`{"minimum":12345678901234567891}`, `12345678901234567890`, false},
		{`{"minimum":12345678901234567890}`, `12345678901234567891`, true},
		{`{"minimum":0.30000000000000000002}`, `0.30000000000000000001`, false},
		{`{"minimum":1}`, `[]`, true},
		{`{"minimum":-1e-10000}`, `-1e10000`, false},
	})
}
