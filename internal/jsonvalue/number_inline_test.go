package jsonvalue

import (
	"math/big"
	"testing"
)

func decodeNum(t *testing.T, s string) *Number {
	t.Helper()
	v, err := DecodeLimited([]byte(s), NumberLimits{})
	if err != nil {
		t.Fatalf("decode %q: %v", s, err)
	}
	n, ok := v.(*Number)
	if !ok {
		t.Fatalf("decode %q: got %T", s, v)
	}
	return n
}

func TestInlineRoundTripsAgainstDecimal(t *testing.T) {
	for _, s := range []string{
		"0", "1", "-1", "12.99", "-12.99", "0.01", "1.0", "100", "1e18", "1e-18",
		"9223372036854775807", "-9223372036854775807", "123e-18", "0.5",
	} {
		n := decodeNum(t, s)
		m, e, ok := n.Inline()
		if !ok {
			t.Fatalf("%s: want small", s)
		}
		neg, bm, be := n.Decimal()
		want := new(big.Int).Set(bm)
		if neg {
			want.Neg(want)
		}
		if be != int64(e) || want.Cmp(big.NewInt(m)) != 0 {
			t.Errorf("%s: Inline=(%d,%d) Decimal=(%v,%d)", s, m, e, want, be)
		}
	}
}

func TestInlineReportsNotSmall(t *testing.T) {
	for _, s := range []string{
		"9223372036854775808", "-9223372036854775808", "12345678901234567890123",
		"1e19", "1e-19", "1e400", "1e-400", "0.1234567890123456789012",
	} {
		if _, _, ok := decodeNum(t, s).Inline(); ok {
			t.Errorf("%s: want not small", s)
		}
	}
}

func TestInlineAllocsZero(t *testing.T) {
	n := decodeNum(t, "12.99")
	var sink int64
	if a := testing.AllocsPerRun(1000, func() { m, _, _ := n.Inline(); sink += m }); a != 0 {
		t.Fatalf("allocs = %v, want 0", a)
	}
	_ = sink
}
