package validation

import (
	"fmt"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
)

// numberValue reads a keyword's value at compile time. A non-number is a
// schema error naming only the keyword. It never builds a big.Rat.
func numberValue(keyword string, v any) (*jsonvalue.Number, error) {
	n, ok := v.(*jsonvalue.Number)
	if !ok {
		return nil, fmt.Errorf("%s: value must be a number", keyword)
	}
	return n, nil
}

// numberInstance returns the instance as a number; ok is false for every
// other JSON type (such instances are valid against numeric keywords).
func numberInstance(instance any) (*jsonvalue.Number, bool) {
	n, ok := instance.(*jsonvalue.Number)
	return n, ok
}
