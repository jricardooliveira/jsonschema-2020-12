package jsonschema

import (
	"errors"

	"github.com/neoface/jsonschema/internal/jsonvalue"
)

// Default number caps, applied when an Options field is 0.
const (
	DefaultMaxNumberDigits   = jsonvalue.DefaultMaxNumberDigits
	DefaultMaxNumberExponent = jsonvalue.DefaultMaxNumberExponent
)

// Options are fixed at compile time. A zero field means the default; a
// negative field is an error.
type Options struct {
	MaxNumberDigits   int
	MaxNumberExponent int
}

func (o Options) limits() (jsonvalue.NumberLimits, error) {
	if o.MaxNumberDigits < 0 || o.MaxNumberExponent < 0 {
		return jsonvalue.NumberLimits{}, errors.New("options: number limits must not be negative")
	}
	l := jsonvalue.NumberLimits{MaxDigits: o.MaxNumberDigits, MaxExponent: o.MaxNumberExponent}
	if l.MaxDigits == 0 {
		l.MaxDigits = DefaultMaxNumberDigits
	}
	if l.MaxExponent == 0 {
		l.MaxExponent = DefaultMaxNumberExponent
	}
	return l, nil
}
