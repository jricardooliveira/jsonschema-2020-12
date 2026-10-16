// Package eval runs a compiled schema against a decoded document.
package eval

import (
	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/keywords"
)

// Evaluate reports whether instance is valid. An error is never valid.
// It only reads n; per-call state lives in the EvalContext.
func Evaluate(n *compile.Node, instance any) (bool, error) {
	if n.Bool != nil {
		return *n.Bool, nil
	}
	ctx := &keywords.EvalContext{}
	valid := true
	for _, it := range n.Items {
		ok, err := it.Keyword.Evaluate(ctx, it.Data, instance)
		if err != nil {
			return false, err
		}
		valid = valid && ok
	}
	return valid, nil
}
