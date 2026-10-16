package jsonvalue

import (
	"math"
	"math/big"
	"strings"
	"sync/atomic"
)

// Number is an exact decimal: value = (-1)^neg * mant * 10^exp, with mant
// free of trailing zeros. Zero has the single form {false, 0, 0, 0}.
// It is immutable: fields are unexported and only the decoder builds one.
type Number struct {
	text string
	neg  bool
	mant big.Int
	exp  int64
	ndig int64 // decimal digits of mant (0 for zero)
}

// ratBuilds counts Rat() expansions; tests read it.
var ratBuilds atomic.Int64

// newNumber builds a Number from token text that already passed checkNumber;
// writtenExp is the exponent that check returned.
func newNumber(text string, writtenExp int64, off int64) (*Number, error) {
	neg, ip, fr, _ := splitNumber(text)
	digits := strings.TrimLeft(ip+fr, "0")
	n := &Number{text: text}
	if digits == "" {
		return n, nil
	}
	trimmed := strings.TrimRight(digits, "0")
	zeros := int64(len(digits) - len(trimmed))
	exp := writtenExp - int64(len(fr))
	// int64 overflow is only possible with absurdly raised caps.
	if exp > math.MaxInt64-zeros-int64(len(trimmed)) || exp < math.MinInt64+int64(len(fr)) {
		return nil, &LimitError{Limit: "number exponent", Max: math.MaxInt64, Offset: off}
	}
	n.exp = exp + zeros
	n.neg = neg
	n.ndig = int64(len(trimmed))
	n.mant.SetString(trimmed, 10)
	return n, nil
}

// IsInteger reports whether the value has a zero fractional part.
func (n *Number) IsInteger() bool { return n.exp >= 0 }

func (n *Number) sign() int {
	switch {
	case n.mant.Sign() == 0:
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
	} else {
		// equal adjusted exponent: |exp difference| = |ndig difference|, bounded by the digit cap.
		x, y := new(big.Int).Set(&a.mant), new(big.Int).Set(&b.mant)
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
	return a.neg == b.neg && a.exp == b.exp && a.mant.Cmp(&b.mant) == 0
}

// Rat returns a fresh exact big.Rat. It is the only place 10 is raised to
// a power for an exponent; hostile-scale callers must not use it.
func (n *Number) Rat() *big.Rat {
	ratBuilds.Add(1)
	r := new(big.Rat).SetInt(&n.mant)
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

func pow10(e int64) *big.Int { return new(big.Int).Exp(big.NewInt(10), big.NewInt(e), nil) }

func cmpInt64(a, b int64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}
