package jsonschema_test

import (
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/speccases"
)

// TestApplicatorSpecCases runs testdata/applicator through the shared runner,
// which also rejects any case without a spec citation.
func TestApplicatorSpecCases(t *testing.T) {
	speccases.Run(t, "testdata/applicator/*.json")
}
