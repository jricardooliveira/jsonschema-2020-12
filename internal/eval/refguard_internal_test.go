package eval

import (
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

// The loop and depth guard fails closed when the per-call record is missing.
func TestEnterNodeFailsClosed(t *testing.T) {
	ctx := keywords.NewEvalContext(DefaultMaxOutputUnits, evalNode)
	ctx.Items = struct{}{}
	n := mustCompile(t, true)
	if _, _, err := enterNode(ctx, n); err == nil {
		t.Error("guard bypassed")
	}
}
