package jsonvalue

// InlineExpLimit bounds the exponent of the small form: |exp| <= InlineExpLimit.
const InlineExpLimit = 18

// Inline returns the number's small decimal form without allocating:
// value = mant * 10^exp, with the mantissa free of trailing zeros. ok is false
// when the mantissa is held in big form or the exponent is outside
// ±InlineExpLimit; callers then use Decimal.
func (n *Number) Inline() (mant int64, exp int32, ok bool) {
	if n.big != nil || n.exp > InlineExpLimit || n.exp < -InlineExpLimit {
		return 0, 0, false
	}
	// mag <= MaxInt64 on the fast path, so the conversion is exact.
	mant = int64(n.mag)
	if n.neg {
		mant = -mant
	}
	return mant, int32(n.exp), true
}
