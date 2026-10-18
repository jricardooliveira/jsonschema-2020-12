package validation

import (
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
)

func mustNum(t *testing.T, s string) *jsonvalue.Number {
	t.Helper()
	v, err := jsonvalue.DecodeLimited([]byte(s), jsonvalue.NumberLimits{})
	if err != nil {
		t.Fatal(err)
	}
	return v.(*jsonvalue.Number)
}

func TestMultipleSmallEdgeCases(t *testing.T) {
	cases := []struct {
		inst, div string
		want      bool
		decided   bool
	}{
		{"12.99", "0.01", true, true},
		{"12.999", "0.01", false, true},
		{"0", "0.01", true, true},
		{"-12.99", "0.01", true, true},
		{"-7", "7", true, true},
		{"0.1", "0.125", false, true},
		{"0.375", "0.125", true, true},
		{"0.0625", "0.0625", true, true},
		{"3125", "3125", true, true},
		{"6250", "3125", true, true},
		{"1", "3125", false, true},
		{"1024", "1024", true, true},
		{"0.00064", "0.00032", true, true},
		{"0.00033", "0.00032", false, true},
		{"1e18", "1e-18", true, true},
		{"1e-18", "1e18", false, false},
		{"1e18", "1e18", true, true},
		{"1e-18", "1e-18", true, true},
		{"9223372036854775807", "1", true, true},
		{"9223372036854775807", "9223372036854775807", true, true},
		{"9223372036854775807", "2", false, true},
		{"-9223372036854775807", "1e-18", true, true},
		{"1", "9223372036854775807e-18", false, true},
	}
	for _, c := range cases {
		inst, div := mustNum(t, c.inst), mustNum(t, c.div)
		m1, e1, ok1 := inst.Inline()
		m2, e2, ok2 := div.Inline()
		if !ok1 || !ok2 {
			t.Fatalf("%s / %s: not small", c.inst, c.div)
		}
		got, decided := multipleSmall(m1, e1, m2, e2)
		if decided != c.decided || (decided && got != c.want) {
			t.Errorf("%s multipleOf %s: got (%v,%v), want (%v,%v)", c.inst, c.div, got, decided, c.want, c.decided)
		}
		if decided {
			c := compileForTest(t, div)
			_, ma, ea := inst.Decimal()
			if big := ma.Sign() == 0 || isMultiple(ma, ea, c); big != got {
				t.Errorf("%s multipleOf %s: small=%v big=%v", inst, div, got, big)
			}
		}
	}
}

func compileForTest(t *testing.T, n *jsonvalue.Number) *multipleOfDivisor {
	t.Helper()
	c, err := compileMultipleOf(nil, n)
	if err != nil {
		t.Fatal(err)
	}
	return c.(*multipleOfDivisor)
}

func TestCompileMultipleOfSmallForm(t *testing.T) {
	if d := compileForTest(t, mustNum(t, "0.01")); !d.small || d.sm != 1 || d.se != -2 {
		t.Errorf("0.01: %+v", d)
	}
	if d := compileForTest(t, mustNum(t, "1e19")); d.small {
		t.Errorf("1e19 must not be small")
	}
}
