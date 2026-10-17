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
	// MaxOutputUnits caps the failure units one Validate may append; 0 means
	// DefaultMaxOutputUnits.
	MaxOutputUnits int
	// MaxAnnotations caps the annotations one Validate may record; 0 means
	// DefaultMaxAnnotations.
	MaxAnnotations int
	// MaxReportBytes caps one rendered report; 0 means DefaultMaxReportBytes.
	MaxReportBytes int
	// MaxDepth, MaxInputBytes, MaxRefChain, MaxSteps and MaxPatternSize: 0
	// means the matching Default constant, negative is an error.
	MaxDepth       int
	MaxInputBytes  int
	MaxRefChain    int
	MaxSteps       int
	MaxPatternSize int
	// FormatAssertion turns format assertion on: a string that fails a known
	// format is invalid. Default off (format is an annotation only).
	FormatAssertion bool
	// BaseURI overrides the base URI of a root schema without $id. It must be
	// absolute with no fragment. See docs/references.md.
	BaseURI string
	// SchemaPath is the file the schema came from; it only gives the default
	// base URI (string work, the path is never opened). Empty means stdin.
	SchemaPath string
}

func (o Options) maxUnits() (int, error) {
	if o.MaxOutputUnits < 0 {
		return 0, errors.New("options: MaxOutputUnits must not be negative")
	}
	if o.MaxOutputUnits == 0 {
		return DefaultMaxOutputUnits, nil
	}
	return o.MaxOutputUnits, nil
}

func (o Options) maxReportBytes() (int, error) {
	if o.MaxReportBytes < 0 {
		return 0, errors.New("options: MaxReportBytes must not be negative")
	}
	if o.MaxReportBytes == 0 {
		return DefaultMaxReportBytes, nil
	}
	return o.MaxReportBytes, nil
}

func (o Options) maxAnnotations() (int, error) {
	if o.MaxAnnotations < 0 {
		return 0, errors.New("options: MaxAnnotations must not be negative")
	}
	if o.MaxAnnotations == 0 {
		return DefaultMaxAnnotations, nil
	}
	return o.MaxAnnotations, nil
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
