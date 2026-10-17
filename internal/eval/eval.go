// Package eval runs a compiled schema against a decoded document.
package eval

import (
	"strings"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/keywords"
)

// Evaluate reports whether instance is valid. An error is never valid.
// It only reads n; per-call state lives in the EvalContext.
func Evaluate(n *compile.Node, instance any) (bool, error) {
	ok, _, err := EvaluateUnits(n, instance, DefaultMaxOutputUnits)
	return ok, err
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
	// core 12.3.2: absoluteKeywordLocation only once a reference was crossed.
	abs := func(loc string) string {
		if !pf.viaRef {
			return ""
		}
		return n.Resource.AbsoluteLocation(loc)
	}
	if n.Bool != nil {
		if *n.Bool {
			return true, nil
		}
		ctx.KeywordLocation = pf.dynamic
		ctx.AbsoluteKeywordLocation = abs(n.Location)
		return false, ctx.Report("schema is false, no value is valid")
	}
	valid := true
	// Two passes: keywords flagged AfterSiblings run last, in schema order.
	for pass := 0; pass < 2; pass++ {
		for _, it := range n.Items {
			if it.Keyword.AfterSiblings != (pass == 1) {
				continue
			}
			ok, err := evalItem(ctx, pf, abs, it, instance)
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
func evalItem(ctx *keywords.EvalContext, pf *pathFrame, abs func(string) string, it compile.Item, instance any) (bool, error) {
	ctx.KeywordLocation = pf.at(it.Keyword.Name, it.Location)
	ctx.AbsoluteKeywordLocation = abs(it.Location)
	ctx.Keyword = it.Keyword.Name
	before := len(ctx.Units())
	rec, _ := ctx.Items.(keywords.ItemRecorder)
	mark := 0
	if rec != nil {
		mark = rec.Mark()
	}
	ok, err := it.Keyword.Evaluate(ctx, it.Data, instance)
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
