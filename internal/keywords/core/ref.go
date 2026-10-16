package core

import (
	"fmt"

	"github.com/neoface/jsonschema/internal/keywords"
)

// Ref is the compiled data of a $ref. Target is filled in once, by load-time
// reference resolution, before the schema is shared; it is read-only after.
type Ref struct {
	Ref      string // the reference as written
	Location string // schema location of the $ref keyword
	Target   any    // compiled node of the resolved target
}

func init() {
	keywords.Register(keywords.Keyword{Name: "$ref", Vocabulary: VocabularyURI, Compile: compileRef, Evaluate: evalRef})
}

func compileRef(ctx *keywords.CompileContext, value any) (any, error) {
	s, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("$ref: must be a string")
	}
	return &Ref{Ref: s, Location: ctx.Location}, nil
}

// evalRef (core 8.2.3.1) applies the resolved target to the same instance;
// sibling keywords are evaluated by the node as usual. An unresolved target is
// an error, never a valid result.
func evalRef(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	r := compiled.(*Ref)
	if r.Target == nil {
		return false, fmt.Errorf("unresolved reference %q at schema location #%s", r.Ref, r.Location)
	}
	ok, units, err := ctx.EvalChild(r.Target, instance, ctx.InstanceLocation)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	ctx.Keep(units)
	return false, ctx.Report("$ref: referenced schema failed")
}
