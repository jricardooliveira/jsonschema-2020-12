package validation_test

// Number paths for uniqueItems: validation spec section 6.4.3.

import "testing"

func TestUniqueItemsNumberPaths(t *testing.T) {
	cases := []struct {
		doc  string
		want bool
	}{
		{"[1,2,3]", true}, {"[1,1.0]", false}, {"[1,1e0]", false}, {"[100,1e2]", false},
		{"[0,-0.0]", false}, {"[0.1,0.10]", false},
		{"[0.1,0.10000000000000001]", true}, {`[1,"1"]`, true},
		{"[9223372036854775807,9223372036854775808]", true},
		{"[9223372036854775807,9223372036854775807.0]", false},
		{"[9223372036854775808,92233720368547758080e-1]", false},
		{"[12345678901234567890,12345678901234567891]", true},
		{"[12345678901234567890,1.2345678901234567890e19]", false},
		{"[[1],[1.0]]", false}, {`[{"a":1},{"a":1.0}]`, false}, {`[{"a":1},{"a":2}]`, true},
	}
	for _, c := range cases {
		if got := npValid(t, `{"uniqueItems":true}`, c.doc); got != c.want {
			t.Errorf("uniqueItems %s: got %v want %v", c.doc, got, c.want)
		}
	}
}
