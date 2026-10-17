package validation_test

// Number paths for maximum (validation spec section 6.2.2). The table is
// shared by the other three bound rules: cmp is sign(doc - bound).

import "testing"

var boundCases = []struct {
	bound, doc string
	cmp        int
}{
	{"9223372036854775807", "9223372036854775807", 0},
	{"9223372036854775807", "9223372036854775808", 1},
	{"9223372036854775808", "9223372036854775807", -1},
	{"9223372036854775807", "9223372036854775807.0", 0},
	{"9223372036854775807", "92233720368547758070e-1", 0},
	{"-9223372036854775808", "-9223372036854775809", -1},
	{"-9223372036854775808", "-9223372036854775808", 0},
	{"0.3", "0.3", 0}, {"0.3", "0.30000000000000001", 1}, {"0.3", "3e-1", 0},
	{"100", "1e2", 0}, {"100", "100.5", 1}, {"1e2", "99.99", -1}, {"5", "5.0", 0},
	{"1", "0.9999999999999999999999", -1},
	{"123456789012345678901234567890", "123456789012345678901234567891", 1},
	{"123456789012345678901234567890", "1234567890123456789012345678900e-1", 0},
	{"123456789012345678901234567890", "9223372036854775807", -1},
}

// runBoundCases checks keyword against boundCases; ok maps cmp to validity.
func runBoundCases(t *testing.T, keyword string, ok func(cmp int) bool) {
	t.Helper()
	// Exponents and digits at the caps (values from type_numberpath_test.go).
	caps := append(boundCases[:0:0], boundCases...)
	caps = append(caps, []struct {
		bound, doc string
		cmp        int
	}{
		{fastAtExp, bigAtExp, 1}, {bigAtExp, fastAtExp, -1}, {fastAtExp, fastAtExp0, 0},
		{"1e" + capExp, nines, -1}, {"1e" + capExp, "10e" + capExpM1, 0},
		{tinyAtExp, "0", -1}, {"0", tinyAtExp, 1}, {nines, fastAtExp, 1},
	}...)
	for _, c := range caps {
		if got := npValid(t, `{"`+keyword+`":`+c.bound+`}`, c.doc); got != ok(c.cmp) {
			t.Errorf("%.30s %s with %.30s: got %v want %v", c.doc, keyword, c.bound, got, ok(c.cmp))
		}
	}
	npLimit(t, `{"`+keyword+`":0}`)
}

func TestMaximumNumberPaths(t *testing.T) {
	runBoundCases(t, "maximum", func(c int) bool { return c <= 0 })
}
