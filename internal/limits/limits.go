// Package limits holds what every package shares about limits: the closed set
// of limit names and the rule that 0 means the default and a negative value
// is rejected. No value switches a limit off.
package limits

import "fmt"

// The closed set of limit names carried by a limit error.
const (
	Depth          = "nesting depth"
	InputSize      = "input size"
	RegistrySize   = "registry size"
	RefChain       = "reference chain"
	Steps          = "evaluation steps"
	NumberDigits   = "number digits"
	NumberExponent = "number exponent"
	PatternSize    = "pattern size"
	OutputUnits    = "output units"
	Annotations    = "annotations"
	ReportBytes    = "report bytes"
)

// Resolve returns def for 0, v for a positive value and an error naming the
// Options field for a negative one.
func Resolve(field string, v, def int) (int, error) {
	switch {
	case v < 0:
		return 0, fmt.Errorf("options: %s must not be negative", field)
	case v == 0:
		return def, nil
	}
	return v, nil
}
