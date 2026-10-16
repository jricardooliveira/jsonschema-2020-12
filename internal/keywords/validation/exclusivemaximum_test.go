package validation_test

// Derived from JSON Schema validation spec section 6.2.3 (exclusiveMaximum).

import (
	"testing"

	"github.com/neoface/jsonschema"
)

func TestExclusiveMaximumBounds(t *testing.T) {
	runBounds(t, []struct {
		schema, doc string
		want        bool
	}{
		{`{"exclusiveMaximum":1}`, `1`, false},
		{`{"exclusiveMaximum":1}`, `1.0`, false},
		{`{"exclusiveMaximum":1}`, `0.9999999999999999999`, true},
		{`{"exclusiveMaximum":12345678901234567890}`, `12345678901234567890`, false},
		{`{"exclusiveMaximum":12345678901234567891}`, `12345678901234567890`, true},
		{`{"exclusiveMaximum":0.30000000000000000002}`, `0.30000000000000000001`, true},
		{`{"exclusiveMaximum":1}`, `"0"`, true},
		{`{"exclusiveMaximum":1}`, `null`, true},
		{`{"exclusiveMaximum":1e10000}`, `1e-10000`, true},
	})
}

func TestExclusiveMaximumSchemaError(t *testing.T) {
	if _, err := jsonschema.Compile([]byte(`{"exclusiveMaximum":"1"}`)); err == nil {
		t.Error("want schema error")
	}
}
