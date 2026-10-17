// Package eval runs a compiled schema against a decoded document.
package eval

import (
	"errors"
	"strings"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

// Evaluate reports whether instance is valid. An error is never valid.
// It only reads n; per-call state lives in the EvalContext.
func Evaluate(n *compile.Node, instance any) (bool, error) {
	ok, _, err := EvaluateUnits(n, instance, DefaultMaxOutputUnits)
	return ok, err
}

// EvaluateValidate is EvaluateLazy for a Validate call: no tree is built, so
// the evaluated-property/item record is skipped when the root node says no
// unevaluated* keyword is reachable (NoUnevaluated, set once at Compile).
// Charging and every limit run as in EvaluateLazy. The switch lives only in
// this call's own state.
// ponytail: duplicates EvaluateLazy's few lines; annotation_record.go was
// outside this task's files. Fold the two together later.
func EvaluateValidate(n *compile.Node, instance any, lim Limits) (Evaluation, error) {
	items := newEvaluatedItems(lim.MaxSteps)
	items.skip = n.NoUnevaluated
	return evaluateOpt(n, instance, lim.MaxUnits, lim.MaxAnnotations, nil, &callState{
		EvaluatedItems: items, active: map[frame]struct{}{},
		maxRefChain: lim.MaxRefChain, maxPath: 4 * lim.MaxDepth,
	}, true)
}

// evalNode evaluates a compiled schema into ctx. Units of failing keywords
// and applicators accumulate in ctx in evaluation order.
func evalNode(ctx *keywords.EvalContext, node any, instance any) (bool, error) {
	n := node.(*compile.Node)
	attachItems(ctx)
	// core 11: unevaluated* see only what this schema object recorded.
	defer func(old int) { ctx.Scope = old }(ctx.Scope)
	if rec, ok := ctx.Items.(keywords.ItemRecorder); ok {
		ctx.Scope = rec.Mark()
	}
	pf, leave, err := enterNode(ctx, n)
	if err != nil {
		return false, err
	}
	defer leave()
	if ctx.Tree != nil {
		ctx.Tree.KeywordLocation = pf.dynamic
	}
	if n.Bool != nil {
		if *n.Bool {
			return true, nil
		}
		ctx.Loc = &lazyLoc{pf: pf, res: n.Resource, static: n.Location}
		return false, ctx.Report("schema is false, no value is valid")
	}
	valid := true
	loc := &lazyLoc{pf: pf, res: n.Resource} // one per node; static is set per keyword
	// Two passes: keywords flagged AfterSiblings run last, in schema order.
	for pass := 0; pass < 2; pass++ {
		for _, it := range n.Items {
			if it.Keyword.AfterSiblings != (pass == 1) {
				continue
			}
			if err := keywordStep(ctx); err != nil { // one step per keyword
				return false, err
			}
			ok, err := evalItem(ctx, pf, loc, it, instance)
			if err != nil {
				return false, err
			}
			if !ok {
				valid = false
			}
		}
	}
	return valid, nil
}

// keywordStep counts one step in the call state enterNode installed; a
// missing state is an error (fail closed), never a panic or a free pass.
func keywordStep(ctx *keywords.EvalContext) error {
	st, ok := ctx.Items.(*callState)
	if !ok {
		return errors.New("internal error: evaluation state missing")
	}
	return st.Step()
}

// lazyLoc is the lazy location of the keyword being evaluated in one node. It
// reads only data fixed at node entry (path frame), the node's resource and
// the keyword's static location; static is set before each keyword and the
// strings are rendered at once, when a unit or annotation is made.
type lazyLoc struct {
	pf     *pathFrame
	res    *compile.Resource
	static string
}

// String renders the keyword location along the evaluation path.
func (k *lazyLoc) String() string {
	return k.pf.dynamic + strings.TrimPrefix(k.static, k.pf.static)
}

// Absolute renders the absolute keyword location (core 12.3.2): empty until
// a reference was crossed.
func (k *lazyLoc) Absolute() string {
	if !k.pf.viaRef {
		return ""
	}
	return k.res.AbsoluteLocation(k.static)
}

// CoveredItems summarises the item annotations recorded at instLoc since
// mark (the schema object's EvalContext.Scope), in plain types for keyword
// packages that cannot import eval (unevaluatedItems, core 11.2): all is true
// when one annotation covers every element, prefix is the largest index
// prefixItems covered (-1 when none) and indices are contains' matches. Each
// entry scanned counts one evaluation step.
func (r *EvaluatedItems) CoveredItems(mark int, instLoc string) (all bool, prefix int, indices []int, err error) {
	prefix = -1
	for _, e := range r.since(mark) {
		if err := r.Step(); err != nil {
			return false, -1, nil, err
		}
		if e.props || e.key.instance != instLoc {
			continue
		}
		switch a := e.ann; {
		case a.All:
			all = true
		case strings.HasSuffix(e.key.keyword, "/prefixItems"):
			prefix = max(prefix, a.Largest)
		default:
			indices = append(indices, a.Indices...)
		}
	}
	return all, prefix, indices, nil
}

// CoveredProperties returns the property names recorded at instLoc since
// mark (the schema object's EvalContext.Scope), for unevaluatedProperties
// (core 11.3). Each entry scanned counts one evaluation step.
func (r *EvaluatedItems) CoveredProperties(mark int, instLoc string) (map[string]bool, error) {
	out := map[string]bool{}
	for _, e := range r.since(mark) {
		if err := r.Step(); err != nil {
			return nil, err
		}
		if e.props && e.key.instance == instLoc {
			for _, n := range e.names {
				out[n] = true
			}
		}
	}
	return out, nil
}

// since returns the entries recorded from mark on.
// ponytail: linear scan of the schema object's own entries, bounded by the
// step cap; index by instance location if deep nesting ever makes it hot.
func (r *EvaluatedItems) since(mark int) []itemEntry {
	if mark < 0 || mark > len(r.entries) {
		return nil
	}
	return r.entries[mark:]
}

// evalItem evaluates one keyword of a schema object and makes sure a failure
// leaves at least one unit. A failing keyword keeps no evaluated-location
// records (core 7.7.1), so a failing 'not' whose subschema passed covers
// nothing for a sibling unevaluated* keyword.
func evalItem(ctx *keywords.EvalContext, pf *pathFrame, loc *lazyLoc, it compile.Item, instance any) (bool, error) {
	loc.static = it.Location
	ctx.Loc = loc
	// Only a reference needs its location now: its target starts there.
	if n := it.Keyword.Name; n == "$ref" || n == "$dynamicRef" {
		pf.at(n, it.Location)
	} else {
		pf.kw = n
	}
	ctx.Keyword = it.Keyword.Name
	before := len(ctx.Units())
	rec, _ := ctx.Items.(keywords.ItemRecorder)
	mark := 0
	if rec != nil {
		mark = rec.Mark()
	}
	var ok bool
	var err error
	if discardsAnnotations(ctx, it.Keyword.Name) {
		leave := enterSink(ctx)
		ok, err = it.Keyword.Evaluate(ctx, it.Data, instance)
		leave()
	} else {
		ok, err = it.Keyword.Evaluate(ctx, it.Data, instance)
	}
	if err != nil {
		return false, err
	}
	if !ok && rec != nil {
		rec.Rollback(mark)
	}
	if !ok && len(ctx.Units()) == before { // keyword gave no unit of its own
		if err := ctx.Report("rule " + it.Keyword.Name + " failed"); err != nil {
			return false, err
		}
	}
	return ok, nil
}
