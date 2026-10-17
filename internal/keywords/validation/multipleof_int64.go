package validation

import (
	"math"
	"math/bits"
)

// multipleSmall decides whether m1*10^e1 is a multiple of m2*10^e2 with int64
// arithmetic only (ADR-004: never floats). m2 must be > 0; both exponents are
// within ±jsonvalue.InlineExpLimit. decided is false when the answer needs
// math/big (an intermediate step would overflow); callers then use isMultiple.
// It never allocates.
func multipleSmall(m1 int64, e1 int32, m2 int64, e2 int32) (multiple, decided bool) {
	if m1 == 0 {
		return true, true
	}
	if m2 <= 0 || m1 == math.MinInt64 { // use the big path
		return false, false
	}
	a := uint64(m1)
	if m1 < 0 {
		a = uint64(-m1)
	}
	b := uint64(m2)
	d := int64(e1) - int64(e2)
	if d >= 0 {
		// m2 | m1*10^d  iff  m2/gcd(m1,m2) has only the factors 2 and 5, at most d of each.
		r := b / gcd(a, b)
		tz := int64(bits.TrailingZeros64(r))
		if tz > d {
			return false, true
		}
		r >>= uint(tz)
		for i := int64(0); i < d && r%5 == 0; i++ {
			r /= 5
		}
		return r == 1, true
	}
	// d < 0: need m2*10^-d | m1.
	t := b
	for i := d; i < 0; i++ {
		hi, lo := bits.Mul64(t, 10)
		if hi != 0 {
			return false, false // overflow: big path decides
		}
		t = lo
	}
	return a%t == 0, true
}

func gcd(a, b uint64) uint64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
