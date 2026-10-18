package validation

import (
	"errors"
	"math/big"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "multipleOf", Vocabulary: VocabularyURI, Compile: compileMultipleOf, Evaluate: evalMultipleOf})
}

// multipleOfDivisor is immutable after compile.
type multipleOfDivisor struct {
	mant   *big.Int // mantissa with all factors of 2 and 5 removed
	v2, v5 int64    // factors of 2 and 5 removed from the mantissa
	exp    int64

	// Small form (int64 mantissa, exponent within ±18); valid only if small.
	small bool
	sm    int64
	se    int32
}

func compileMultipleOf(_ *keywords.CompileContext, value any) (any, error) {
	n, err := numberValue("multipleOf", value)
	if err != nil {
		return nil, err
	}
	neg, m, exp := n.Decimal()
	if neg || m.Sign() == 0 {
		return nil, errors.New("multipleOf: value must be greater than 0")
	}
	d := &multipleOfDivisor{exp: exp}
	d.sm, d.se, d.small = n.Inline()
	d.v2 = int64(m.TrailingZeroBits())
	m.Rsh(m, uint(d.v2))
	d.v5 = stripFives(m)
	d.mant = m
	return d, nil
}

// stripFives divides 5 out of m (in place) and returns how many times.
func stripFives(m *big.Int) int64 {
	var c int64
	q, r, five := new(big.Int), new(big.Int), big.NewInt(5)
	for {
		q.QuoRem(m, five, r)
		if r.Sign() != 0 {
			return c
		}
		m.Set(q)
		c++
	}
}

// hasFactors reports whether m has at least need factors f. need is bounded
// by m's bit length, so work depends on digit counts only.
func hasFactors(m *big.Int, f, need int64) bool {
	m = new(big.Int).Set(m)
	q, r, bf := new(big.Int), new(big.Int), big.NewInt(f)
	for i := int64(0); i < need; i++ {
		q.QuoRem(m, bf, r)
		if r.Sign() != 0 {
			return false
		}
		m.Set(q)
	}
	return true
}

func evalMultipleOf(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	n, ok := numberInstance(instance)
	if !ok {
		return true, nil
	}
	div := compiled.(*multipleOfDivisor)
	if m, e, ok := n.Inline(); ok && div.small {
		if multiple, decided := multipleSmall(m, e, div.sm, div.se); decided {
			if multiple {
				return true, nil
			}
			return false, ctx.Report("value is not a multiple of the multipleOf value")
		}
	}
	_, ma, ea := n.Decimal()
	if ma.Sign() == 0 || isMultiple(ma, ea, div) {
		return true, nil
	}
	return false, ctx.Report("value is not a multiple of the multipleOf value")
}

// isMultiple: a = ma*10^ea over b = mb*10^eb is an integer iff rb | ma,
// v2(ma)+d >= v2(mb) and v5(ma)+d >= v5(mb), with d = ea-eb.
func isMultiple(ma *big.Int, ea int64, b *multipleOfDivisor) bool {
	d := new(big.Int).Sub(big.NewInt(ea), big.NewInt(b.exp)) // cannot overflow
	if new(big.Int).Mod(ma, b.mant).Sign() != 0 {
		return false
	}
	for _, c := range []struct{ f, vb int64 }{{2, b.v2}, {5, b.v5}} {
		need := new(big.Int).Sub(big.NewInt(c.vb), d)
		if need.Sign() <= 0 {
			continue
		}
		if !need.IsInt64() || need.Int64() > int64(ma.BitLen()) || !hasFactors(ma, c.f, need.Int64()) {
			return false
		}
	}
	return true
}
