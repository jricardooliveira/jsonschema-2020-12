package validation_test

// Derived from JSON Schema validation spec section 6.2.5 (exclusiveMinimum).

import (
	"testing"

	"github.com/neoface/jsonschema"
)

func TestExclusiveMinimumBounds(t *testing.T) {
	runBounds(t, []struct {
		schema, doc string
		want        bool
	}{
		{`{"exclusiveMinimum":1}`, `1`, false},
		{`{"exclusiveMinimum":1}`, `1.0`, false},
		{`{"exclusiveMinimum":1}`, `1.0000000000000000001`, true},
		{`{"exclusiveMinimum":12345678901234567890}`, `12345678901234567890`, false},
		{`{"exclusiveMinimum":12345678901234567890}`, `12345678901234567891`, true},
		{`{"exclusiveMinimum":0.30000000000000000001}`, `0.30000000000000000002`, true},
		{`{"exclusiveMinimum":1}`, `"0"`, true},
		{`{"exclusiveMinimum":1}`, `{}`, true},
		{`{"exclusiveMinimum":-1e-10000}`, `-1e10000`, false},
	})
}

func TestExclusiveMinimumSchemaError(t *testing.T) {
	if _, err := jsonschema.Compile([]byte(`{"exclusiveMinimum":null}`)); err == nil {
		t.Error("want schema error")
	}
}
