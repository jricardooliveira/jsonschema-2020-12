package jsonvalue

import (
	"fmt"
	"math/big"
	"runtime"
	"strings"
	"testing"
)

func num(t *testing.T, s string) *Number {
	t.Helper()
	v, err := DecodeLimited([]byte(s), NumberLimits{})
	if err != nil {
		t.Fatalf("%s: %v", s, err)
	}
	return v.(*Number)
}

func TestNumberCmpAndIntegers(t *testing.T) {
	nines := strings.Repeat("9", DefaultMaxNumberDigits)
	e := fmt.Sprint(DefaultMaxNumberExponent)
	cases := []struct {
		a, b string
		want int
	}{
		{"1e" + e, nines, 1},
		{"1e-" + e, "1e-" + fmt.Sprint(DefaultMaxNumberExponent-1), -1},
		{"-1e" + e, "-1", -1},
		{"1e-" + e, "0", 1},
		{"0", "-1e-" + e, 1},
		{"0", "-0", 0}, {"0.0", "0e5", 0}, {"-0", "0e5", 0},
		{"1", "1.0", 0}, {"1", "1e0", 0}, {"1", "10e-1", 0}, {"100", "1e2", 0},
		{"0.1", "0.10000000000000001", -1},
		{"12345678901234567890", "12345678901234567891", -1},
		{"1e" + e, "10e" + fmt.Sprint(DefaultMaxNumberExponent-1), 0},
		{"0.1", "1e-1", 0},
		{"-2", "-1", -1},
	}
	before := ratBuilds.Load()
	for _, c := range cases {
		a, b := num(t, c.a), num(t, c.b)
		if got := Cmp(a, b); got != c.want {
			t.Errorf("Cmp(%.20s, %.20s) = %d, want %d", c.a, c.b, got, c.want)
		}
		if Cmp(b, a) != -c.want {
			t.Errorf("Cmp not antisymmetric for %.20s, %.20s", c.a, c.b)
		}
		if numEqual(a, b) != (c.want == 0) {
			t.Errorf("numEqual(%.20s, %.20s) wrong", c.a, c.b)
		}
	}
	for _, s := range []string{"1.0", "1e2", "-0", "100e-2", "12345678901234567890"} {
		if !num(t, s).IsInteger() {
			t.Errorf("%s should be an integer", s)
		}
	}
	for _, s := range []string{"1e-" + e, "0.1", "1.5"} {
		if num(t, s).IsInteger() {
			t.Errorf("%s should not be an integer", s)
		}
	}
	if got := num(t, "12345678901234567890").Rat().String(); got != "12345678901234567890/1" {
		t.Errorf("big integer lost precision: %s", got)
	}
	if ratBuilds.Load() != before+1 {
		t.Errorf("Cmp/numEqual/IsInteger built a Rat")
	}
}

func TestNumberCapScaleNeverBuildsRat(t *testing.T) {
	e := fmt.Sprint(DefaultMaxNumberExponent)
	before := ratBuilds.Load()
	big1, small, digits := num(t, "1e"+e), num(t, "1e-"+e), num(t, strings.Repeat("9", DefaultMaxNumberDigits))
	alt := num(t, "10e"+fmt.Sprint(DefaultMaxNumberExponent-1))
	for i := 0; i < 10000; i++ {
		Cmp(big1, small)
		Cmp(digits, big1)
		numEqual(big1, alt)
		big1.IsInteger()
		small.IsInteger()
	}
	if ratBuilds.Load() != before {
		t.Fatal("cap-scale operations built a Rat")
	}
	if n := big1.ndig + int64(len(big1.text)); n > int64(len(big1.text))*2 {
		t.Errorf("retained size %d exceeds text size", n)
	}
}

func TestNumberRatMatchesSetString(t *testing.T) {
	for _, s := range []string{"0", "-0", "1", "-1.5", "0.1", "100", "1e5", "1.25e-3", "-12345678901234567890.5"} {
		before := ratBuilds.Load()
		want, _ := new(big.Rat).SetString(s)
		if got := num(t, s).Rat(); got.Cmp(want) != 0 {
			t.Errorf("%s: got %v want %v", s, got, want)
		}
		if ratBuilds.Load() != before+1 {
			t.Errorf("%s: Rat did not count", s)
		}
	}
}

func TestNumberCap(t *testing.T) {
	d, e := DefaultMaxNumberDigits, DefaultMaxNumberExponent
	ok := []string{
		strings.Repeat("1", d), "0." + strings.Repeat("0", d-2) + "1", "1" + strings.Repeat("0", d-1),
		fmt.Sprintf("1e%d", e), fmt.Sprintf("1e-%d", e), fmt.Sprintf("1e+%d", e),
		fmt.Sprintf("1e%0*d", e/2, 5), // leading zeros within the cap
	}
	over := []string{
		strings.Repeat("1", d+1), "0." + strings.Repeat("0", d-1) + "1", "1" + strings.Repeat("0", d),
		"0." + strings.Repeat("0", d) + "1",
		fmt.Sprintf("1e%d", e+1), fmt.Sprintf("1e-%d", e+1), "1e999999999", "1e-999999999",
		"1e" + strings.Repeat("9", 25), "1e-" + strings.Repeat("9", 25),
	}
	for _, s := range ok {
		if _, err := DecodeLimited([]byte(s), NumberLimits{}); err != nil {
			t.Errorf("%.30s should decode: %v", s, err)
		}
		if _, err := Decode([]byte(s)); err != nil {
			t.Errorf("Decode %.30s: %v", s, err)
		}
	}
	for _, s := range over {
		_, err := DecodeLimited([]byte(s), NumberLimits{})
		le, isLimit := err.(*LimitError)
		if !isLimit {
			t.Errorf("%.30s: want *LimitError, got %v", s, err)
			continue
		}
		if strings.Contains(le.Error(), "9999") || strings.Contains(le.Error(), "111") {
			t.Errorf("limit error quotes number: %v", le)
		}
		if _, err := Decode([]byte(s)); err == nil {
			t.Errorf("Decode %.30s should be capped", s)
		}
	}
	// raised cap accepts it
	if _, err := DecodeLimited([]byte(fmt.Sprintf("1e%d", e+1)), NumberLimits{MaxExponent: e + 1}); err != nil {
		t.Error(err)
	}
	// zero field means default, never unlimited
	if _, err := DecodeLimited([]byte(fmt.Sprintf("1e%d", e+1)), NumberLimits{MaxDigits: 5}); err == nil {
		t.Error("zero MaxExponent must not mean unlimited")
	}
}

func TestNumberManyTokensStayLinear(t *testing.T) {
	if testing.Short() {
		t.Skip("allocation test")
	}
	const bytesPerInputByte = 100 // generous allocation bound per input byte
	tok := fmt.Sprintf("1e%d,1e-%d,", DefaultMaxNumberExponent, DefaultMaxNumberExponent)
	input := "[" + strings.Repeat(tok, 100*DefaultMaxNumberDigits/2) + "0]"
	before := ratBuilds.Load()
	var m0, m1 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m0)
	v, err := DecodeLimited([]byte(input), NumberLimits{})
	runtime.ReadMemStats(&m1)
	if err != nil || len(v.([]any)) != 100*DefaultMaxNumberDigits+1 {
		t.Fatalf("decode failed: %v", err)
	}
	if ratBuilds.Load() != before {
		t.Error("decode built a Rat")
	}
	if grew := m1.TotalAlloc - m0.TotalAlloc; grew > uint64(bytesPerInputByte*len(input)) {
		t.Errorf("allocated %d bytes for %d input bytes", grew, len(input))
	}
}
