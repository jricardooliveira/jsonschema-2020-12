package validation_test

// Number paths for multipleOf: validation spec section 6.2.1.

import "testing"

func TestMultipleOfNumberPaths(t *testing.T) {
	cases := []struct {
		div, doc string
		want     bool
	}{
		{"0.1", "0.3", true}, {"0.1", "0.7", true}, {"0.1", "0.35", false},
		{"0.1", "0.30000000000000001", false},
		{"1", "1.0", true}, {"0.01", "0.99", true}, {"0.01", "4.35", true},
		{"0.5", "7.5", true}, {"2.5", "7.5", true}, {"0.0001", "0.0003", true},
		{"1e2", "1000", true}, {"1e2", "150", false},
		{"3", "0", true}, {"3", "0.0", true},
		// int64 boundary: 2^63-1 is divisible by 7 but not by 3; 2^63 by 2 not 3.
		{"7", "9223372036854775807", true}, {"3", "9223372036854775807", false},
		{"3", "9223372036854775806", true},
		{"2", "9223372036854775808", true}, {"3", "9223372036854775808", false},
		{"0.1", "1e30", true}, {"0.1", "9223372036854775808.1", true}, {"0.1", "9223372036854775808.05", false},
		{"10", "123456789012345678901234567890", true},
		{"20", "123456789012345678901234567890", false},
		{"123456789012345678901234567890", "246913578024691357802469135780", true},
		{"9223372036854775808", "18446744073709551616", true},
		{"9223372036854775807", "9223372036854775808", false},
	}
	for _, c := range cases {
		if got := npValid(t, `{"multipleOf":`+c.div+`}`, c.doc); got != c.want {
			t.Errorf("%s multipleOf %s: got %v want %v", c.doc, c.div, got, c.want)
		}
	}
}
