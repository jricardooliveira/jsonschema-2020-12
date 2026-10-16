package jsonvalue

import "strconv"

// Canonical returns the normalised value as text: sign, mantissa digits and
// exponent. Equal numbers give equal text (1, 1.0 and 1e0 all match). It reads
// the stored fields only and never expands a power of ten.
func (n *Number) Canonical() string {
	s := "+"
	if n.neg {
		s = "-"
	}
	return s + n.mant.Text(10) + "e" + strconv.FormatInt(n.exp, 10)
}

// String returns the number as written in the source. It is used for schema
// values only in messages (never for instance values).
func (n *Number) String() string { return n.text }
