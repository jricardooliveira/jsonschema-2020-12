package applicator

import (
	"fmt"

	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "if", Vocabulary: VocabularyURI, Compile: compileIf, Evaluate: evalIf})
}

func compileIf(ctx *keywords.CompileContext, value any) (any, error) {
	n, err := ctx.Sub(value, "")
	if err != nil {
		return nil, fmt.Errorf("if: %w", err)
	}
	return n, nil
}

// evalIf (core 10.2.2.1) never fails a document and reports nothing.
// ponytail: 'then' and 'else' evaluate the sibling 'if' themselves, because
// keywords cannot see each other's results; 'if' is only compiled here (so a
// malformed value is a schema error). Upgrade path: a per-call result record
// if 'if' annotations must be kept for unevaluated* when no branch exists.
func evalIf(*keywords.EvalContext, any, any) (bool, error) { return true, nil }

// branch is the compiled form of then/else: the sibling 'if' (nil when there
// is none) and the branch's own subschema.
type branch struct {
	cond, node any
}

func compileBranch(name string, ctx *keywords.CompileContext, value any) (any, error) {
	n, err := ctx.Sub(value, "")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	b := &branch{node: n}
	if c, ok := ctx.Siblings["if"]; ok {
		if b.cond, err = ctx.Sub(c, ""); err != nil {
			return nil, fmt.Errorf("%s: if: %w", name, err)
		}
	}
	return b, nil
}

// evalBranch applies the branch when 'if' evaluates to wantIf. The 'if' units
// are dropped (still counted); the branch's units are kept on failure.
func evalBranch(name string, wantIf bool, ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	b := compiled.(*branch)
	if b.cond == nil {
		return true, nil
	}
	ok, _, err := ctx.EvalChild(b.cond, instance, ctx.InstanceLocation)
	if err != nil {
		return false, err
	}
	if ok != wantIf {
		return true, nil
	}
	ok, units, err := ctx.EvalChild(b.node, instance, ctx.InstanceLocation)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	ctx.Keep(units)
	return false, ctx.Report(name + ": subschema did not match")
}
