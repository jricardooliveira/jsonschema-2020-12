package jsonvalue

import (
	"math"
	"math/big"
	"strconv"
	"strings"
	"sync/atomic"
)

// Number is an exact decimal: value = (-1)^neg * mag * 10^exp, with the
// mantissa free of trailing zeros. The mantissa is the inline mag field while
// it fits int64; only wider mantissas use big (never both: a big that fits
// int64 is demoted). Zero has the single form {false, 0, nil, 0, 0}.
// It is immutable: fields are unexported and only the decoder builds one.
type Number struct {
	text string
	neg  bool
	mag  uint64   // magnitude on the fast path; 0 when big != nil
	big  *big.Int // magnitude when it exceeds math.MaxInt64, else nil
	exp  int64
	ndig int64 // decimal digits of the mantissa (0 for zero)
}

// ratBuilds counts Rat() expansions; tests read it.
var ratBuilds atomic.Int64

var pow10tab = func() (t [20]uint64) {
	t[0] = 1
	for i := 1; i < len(t); i++ {
		t[i] = t[i-1] * 10
	}
	return
}()

// newNumber builds a Number from token text that already passed checkNumber;
// writtenExp is the exponent that check returned.
func newNumber(text string, writtenExp int64, off int64) (*Number, error) {
	neg, ip, fr, _ := splitNumber(text)
	// Fast scan: accumulate up to 19 significant digits with no allocation.
	var v uint64
	var total, tz int64
scan:
	for _, part := range [2]string{ip, fr} {
		for i := 0; i < len(part); i++ {
			c := part[i]
			if total == 0 && c == '0' {
				continue
			}
			if total == 19 {
				total++
				break scan
			}
			v = v*10 + uint64(c-'0')
			total++
			if c == '0' {
				tz++
			} else {
				tz = 0
			}
		}
	}
	n := &Number{text: text}
	if total == 0 {
		return n, nil
	}
	var trimmed string
	if total <= 19 && v/pow10tab[tz] <= math.MaxInt64 {
		n.mag = v / pow10tab[tz]
		n.ndig = total - tz
	} else {
		digits := strings.TrimLeft(ip+fr, "0")
		trimmed = strings.TrimRight(digits, "0")
		tz = int64(len(digits) - len(trimmed))
		n.ndig = int64(len(trimmed))
	}
	// int64 overflow is only possible with absurdly raised caps. Check
	// before subtracting, and keep adj = ndig + exp representable.
	if writtenExp < math.MinInt64+int64(len(fr)) {
		return nil, &LimitError{Limit: "number exponent", Max: math.MaxInt64, Offset: off}
	}
	exp := writtenExp - int64(len(fr))
	if exp > math.MaxInt64-tz-n.ndig {
		return nil, &LimitError{Limit: "number exponent", Max: math.MaxInt64, Offset: off}
	}
	n.exp = exp + tz
	// exp == MinInt64 would make -exp overflow in Rat(): keep it unrepresentable.
	if n.exp < -math.MaxInt64 {
		return nil, &LimitError{Limit: "number exponent", Max: math.MaxInt64, Offset: off}
	}
	n.neg = neg
	if trimmed != "" {
		if u, err := strconv.ParseUint(trimmed, 10, 64); err == nil && u <= math.MaxInt64 {
			n.mag = u // demote: fits the fast form
		} else {
			n.big, _ = new(big.Int).SetString(trimmed, 10)
		}
	}
	return n, nil
}

// magnitude returns a fresh copy of the mantissa magnitude.
func (n *Number) magnitude() *big.Int {
	if n.big != nil {
		return new(big.Int).Set(n.big)
	}
	return new(big.Int).SetUint64(n.mag)
}

// Decimal returns the sign, a fresh copy of the mantissa and the exponent:
// value = (-1)^neg * mant * 10^exp. It never raises 10 to a power.
func (n *Number) Decimal() (neg bool, mant *big.Int, exp int64) {
	return n.neg, n.magnitude(), n.exp
}

// IsInteger reports whether the value has a zero fractional part.
func (n *Number) IsInteger() bool { return n.exp >= 0 }

func (n *Number) sign() int {
	switch {
	case n.big == nil && n.mag == 0:
		return 0
	case n.neg:
		return -1
	}
	return 1
}

// Cmp compares by mathematical value: -1, 0 or +1.
func Cmp(a, b *Number) int {
	sa, sb := a.sign(), b.sign()
	if sa != sb || sa == 0 {
		return cmpInt64(int64(sa), int64(sb))
	}
	var c int
	if aa, ab := a.ndig+a.exp, b.ndig+b.exp; aa != ab {
		c = cmpInt64(aa, ab)
	} else if a.big == nil && b.big == nil {
		// equal adjusted exponent: |exp difference| = |ndig difference| <= 19, and
		// the aligned value has at most 19 digits, so it fits uint64.
		x, y := a.mag, b.mag
		if d := a.exp - b.exp; d > 0 {
			x *= pow10tab[d]
		} else if d < 0 {
			y *= pow10tab[-d]
		}
		switch {
		case x < y:
			c = -1
		case x > y:
			c = 1
		}
	} else {
		// |exp difference| = |ndig difference|, bounded by the digit cap.
		x, y := a.magnitude(), b.magnitude()
		if d := a.exp - b.exp; d > 0 {
			x.Mul(x, pow10(d))
		} else if d < 0 {
			y.Mul(y, pow10(-d))
		}
		c = x.Cmp(y)
	}
	if sa < 0 {
		c = -c
	}
	return c
}

// numEqual is exact equality without allocating (used by Equal).
func numEqual(a, b *Number) bool {
	if a.text == b.text {
		return true
	}
	if a.neg != b.neg || a.exp != b.exp {
		return false
	}
	if a.big == nil || b.big == nil {
		return a.big == b.big && a.mag == b.mag // canonical: fast never equals big
	}
	return a.big.Cmp(b.big) == 0
}

// Rat returns a fresh exact big.Rat. It is the only place 10 is raised to
// a power for an exponent; hostile-scale callers must not use it.
func (n *Number) Rat() *big.Rat {
	ratBuilds.Add(1)
	r := new(big.Rat).SetInt(n.magnitude())
	if n.exp > 0 {
		r.Mul(r, new(big.Rat).SetInt(pow10(n.exp)))
	} else if n.exp < 0 {
		r.Quo(r, new(big.Rat).SetInt(pow10(-n.exp)))
	}
	if n.neg {
		r.Neg(r)
	}
	return r
}

// pow10 returns 10^e. A negative e is an invariant violation (big.Int.Exp
// would silently return 1), so it panics instead of giving a wrong value.
func pow10(e int64) *big.Int {
	if e < 0 {
		panic("jsonvalue: negative power of ten")
	}
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(e), nil)
}

func cmpInt64(a, b int64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}
