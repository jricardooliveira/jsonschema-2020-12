package jsonvalue

import (
	"math"
	"math/big"
	"strconv"
	"testing"
)

func TestNumberFastBigRepresentation(t *testing.T) {
	maxI := strconv.FormatInt(math.MaxInt64, 10)
	cases := []struct {
		in        string
		wantBig   bool
		canonical string
	}{
		{"0", false, "+0e0"}, {"-0.0", false, "+0e0"},
		{"1", false, "+1e0"}, {"1.0", false, "+1e0"}, {"1e0", false, "+1e0"},
		{"100", false, "+1e2"}, {"1e2", false, "+1e2"},
		{"0.99", false, "+99e-2"}, {"0.01", false, "+1e-2"}, {"-0.10", false, "-1e-1"},
		{maxI, false, "+" + maxI + "e0"},
		{"9223372036854775808", true, "+9223372036854775808e0"},
		{"92233720368547758070", true, "+9223372036854775807e1"}, // trailing zero folds, demotes
		{"9223372036854775807.0", false, "+" + maxI + "e0"},
		{"-9223372036854775809", true, "-9223372036854775809e0"},
		{"12345678901234567890", true, "+1234567890123456789e1"},
		{"0.00000000000000000000000000001", false, "+1e-29"},
	}
	for _, c := range cases {
		n := num(t, c.in)
		if got := n.Canonical(); got != c.canonical {
			t.Errorf("%s: canonical %s want %s", c.in, got, c.canonical)
		}
		if isBig := n.big != nil; isBig && !c.wantBig && c.in != "92233720368547758070" {
			t.Errorf("%s: unexpected big", c.in)
		}
		if n.String() != c.in {
			t.Errorf("%s: text changed", c.in)
		}
	}
	// 19 digits that fit stay fast, 19 digits that do not go big.
	if num(t, maxI).big != nil || num(t, "9223372036854775808").big == nil {
		t.Error("int64 boundary misplaced")
	}
	if n := num(t, "92233720368547758070"); n.big != nil || n.mag != math.MaxInt64 || n.exp != 1 {
		t.Errorf("trailing zeros not folded and demoted: %+v", n)
	}
}

func TestNumberFastBigMixedExact(t *testing.T) {
	maxI := strconv.FormatInt(math.MaxInt64, 10)
	eq := [][2]string{{"1", "1.0"}, {"1", "1e0"}, {"100", "1e2"}, {maxI, maxI + ".0"},
		{"9223372036854775808", "92233720368547758080e-1"}}
	for _, p := range eq {
		a, b := num(t, p[0]), num(t, p[1])
		if !numEqual(a, b) || Cmp(a, b) != 0 {
			t.Errorf("%s != %s", p[0], p[1])
		}
	}
	lt := [][2]string{{maxI, "9223372036854775808"}, {"9223372036854775806", maxI},
		{"0.3", "0.30000000000000001"}, {"-9223372036854775808", "-9223372036854775807"},
		{"1.5", "15.1e-1"}, {"9223372036854775807", "92233720368547758071e-1"}}
	for _, p := range lt {
		a, b := num(t, p[0]), num(t, p[1])
		if Cmp(a, b) != -1 || Cmp(b, a) != 1 || numEqual(a, b) {
			t.Errorf("%s should be < %s", p[0], p[1])
		}
	}
}

func TestNumberFastPathNoBigAndDecimal(t *testing.T) {
	n := num(t, "-12.50")
	if n.big != nil {
		t.Fatal("fast number holds a big.Int")
	}
	neg, m, e := n.Decimal()
	if !neg || m.Cmp(big.NewInt(125)) != 0 || e != -1 {
		t.Errorf("Decimal = %v %v %d", neg, m, e)
	}
	m.SetInt64(7) // fresh copy
	if _, m2, _ := n.Decimal(); m2.Int64() != 125 {
		t.Error("Decimal returned shared state")
	}
	if got := n.Rat().String(); got != "-25/2" {
		t.Errorf("Rat = %s", got)
	}
	if n.IsInteger() || !num(t, "1.0").IsInteger() {
		t.Error("IsInteger wrong")
	}
}
