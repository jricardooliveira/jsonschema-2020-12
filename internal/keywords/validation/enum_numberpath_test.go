package validation_test

// Number paths for enum: validation spec section 6.1.2 (deep JSON equality).

import "testing"

// equalCases: value, doc, equal. Shared by enum and const.
var equalCases = []struct {
	val, doc string
	want     bool
}{
	{"1", "1.0", true}, {"1", "1e0", true}, {"1", "1.5", false},
	{"0.1", "0.10", true}, {"0.1", "1e-1", true}, {"0.1", "0.10000000000000001", false},
	{"100", "1e2", true}, {"0", "-0.0", true},
	{"9223372036854775807", "9223372036854775807.0", true},
	{"9223372036854775807", "9223372036854775808", false},
	{"9223372036854775808", "92233720368547758080e-1", true},
	{"12345678901234567890", "1.2345678901234567890e19", true},
	{"12345678901234567890", "12345678901234567891", false},
	{`"1"`, "1", false},
	{`[1,{"a":2.0}]`, `[1.0,{"a":2}]`, true},
	{`[1,{"a":2.0}]`, `[1.0,{"a":2.1}]`, false},
}

// equalCapCases: equalCases at the exponent and digit caps (values from
// type_numberpath_test.go).
func equalCapCases() []struct {
	val, doc string
	want     bool
} {
	return []struct {
		val, doc string
		want     bool
	}{
		{"1e" + capExp, "10e" + capExpM1, true}, {fastAtExp, fastAtExp0, true},
		{fastAtExp, bigAtExp, false}, {nines, "9." + nines[1:] + "e" + capExpM1, true},
		{nines, "1e" + capExp, false}, {tinyAtExp, "0", false}, {tinyAtExp, "0.1e-" + capExpM1, true},
	}
}

func TestEnumNumberPaths(t *testing.T) {
	for _, c := range append(equalCases, equalCapCases()...) {
		if got := npValid(t, `{"enum":[`+c.val+`]}`, c.doc); got != c.want {
			t.Errorf("enum [%.30s] doc %.30s: got %v want %v", c.val, c.doc, got, c.want)
		}
	}
	npLimit(t, `{"enum":[1]}`)
}
