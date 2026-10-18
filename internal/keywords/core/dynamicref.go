package core

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

// dynamicResolver is implemented by the per-call evaluation state. It maps the
// statically resolved target to the outermost matching $dynamicAnchor in the
// dynamic scope (core 8.2.3.2), or returns initial unchanged.
type dynamicResolver interface {
	ResolveDynamic(initial any, name string) any
}

func init() {
	keywords.Register(keywords.Keyword{Name: "$dynamicRef", Vocabulary: VocabularyURI, Compile: compileDynamicRef, Evaluate: evalDynamicRef})
}

// compileDynamicRef returns a *Ref: load-time resolution treats the initial
// target of $dynamicRef exactly like a $ref target, so an unresolvable one is
// a schema error.
func compileDynamicRef(ctx *keywords.CompileContext, value any) (any, error) {
	s, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("$dynamicRef: must be a string")
	}
	return &Ref{Ref: s, Location: ctx.Location}, nil
}

// plainName returns the fragment of ref when it is a plain name (not empty,
// not a JSON Pointer).
func plainName(ref string) (string, bool) {
	_, frag, ok := strings.Cut(ref, "#")
	if !ok || frag == "" || frag[0] == '/' {
		return "", false
	}
	name, err := url.PathUnescape(frag)
	return name, err == nil
}

// evalDynamicRef applies the initial target, or, when that target carries a
// matching $dynamicAnchor, the outermost one in the dynamic scope. Re-entry
// is checked by the evaluator on the resolved node, so loops are errors.
func evalDynamicRef(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	r := compiled.(*Ref)
	if r.Target == nil {
		return false, fmt.Errorf("unresolved reference %q at schema location #%s", r.Ref, r.Location)
	}
	target := r.Target
	if name, ok := plainName(r.Ref); ok {
		if dr, ok := ctx.Items.(dynamicResolver); ok {
			target = dr.ResolveDynamic(target, name)
		}
	}
	ok, units, err := ctx.EvalChild(target, instance, ctx.InstanceLocation)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	ctx.Keep(units)
	return false, ctx.Report("$dynamicRef: referenced schema failed")
}
