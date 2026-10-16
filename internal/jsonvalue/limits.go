package jsonvalue

import (
	"strconv"
)

// Default number caps: far above real-world values.
const (
	DefaultMaxNumberDigits   = 10000
	DefaultMaxNumberExponent = 10000
)

// NumberLimits caps number size. A zero or negative field is never read as
// "no limit": it falls back to the default.
type NumberLimits struct {
	MaxDigits   int
	MaxExponent int
}

func (l NumberLimits) resolved() NumberLimits {
	if l.MaxDigits <= 0 {
		l.MaxDigits = DefaultMaxNumberDigits
	}
	if l.MaxExponent <= 0 {
		l.MaxExponent = DefaultMaxNumberExponent
	}
	return l
}

// splitNumber splits valid JSON number text into sign, integer digits,
// fraction digits and the exponent text (sign included, may be empty).
func splitNumber(s string) (neg bool, intPart, frac, exp string) {
	if s != "" && s[0] == '-' {
		neg, s = true, s[1:]
	}
	for i := 0; i < len(s); i++ {
		if s[i] == 'e' || s[i] == 'E' {
			exp, s = s[i+1:], s[:i]
			break
		}
	}
	intPart = s
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			intPart, frac = s[:i], s[i+1:]
			break
		}
	}
	return
}

// checkNumber enforces the caps on the raw token text, before any
// strconv or math/big work. Every written digit counts, zeros included.
// It returns the written exponent.
func checkNumber(s string, lim NumberLimits, off int64) (int64, error) {
	_, ip, fr, ex := splitNumber(s)
	if len(ip)+len(fr) > lim.MaxDigits {
		return 0, &LimitError{Limit: "number digits", Max: lim.MaxDigits, Offset: off}
	}
	if ex == "" {
		return 0, nil
	}
	eneg := false
	if ex[0] == '+' || ex[0] == '-' {
		eneg, ex = ex[0] == '-', ex[1:]
	}
	for len(ex) > 1 && ex[0] == '0' {
		ex = ex[1:]
	}
	if len(ex) > len(strconv.Itoa(lim.MaxExponent)) {
		return 0, &LimitError{Limit: "number exponent", Max: lim.MaxExponent, Offset: off}
	}
	e, err := strconv.ParseInt(ex, 10, 64)
	if err != nil || e > int64(lim.MaxExponent) {
		return 0, &LimitError{Limit: "number exponent", Max: lim.MaxExponent, Offset: off}
	}
	if eneg {
		e = -e
	}
	return e, nil
}
