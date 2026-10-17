package unevaluated

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/neoface/jsonschema/internal/keywords"
	"github.com/neoface/jsonschema/internal/keywords/validation"
)

func init() {
	keywords.Register(keywords.Keyword{
		Name: "unevaluatedItems", Vocabulary: VocabularyURI,
		Compile: compileUnevaluatedItems, Evaluate: evalUnevaluatedItems,
		AfterSiblings: true, // core 11.2: sees every sibling's annotations
	})
}

// itemLog reads the per-call item annotations (an *eval.EvaluatedItems). It
// uses plain types: importing eval here would make an import cycle.
type itemLog interface {
	CoveredItems(instLoc string) (all bool, prefix int, indices []int)
}

func compileUnevaluatedItems(ctx *keywords.CompileContext, value any) (any, error) {
	n, err := ctx.Sub(value, "")
	if err != nil {
		return nil, fmt.Errorf("unevaluatedItems: %w", err)
	}
	return n, nil
}

// evalUnevaluatedItems applies the subschema to every element that no
// successful evaluation at this instance location covered (core 11.2). Each
// failing element yields its own units at its own instance location.
func evalUnevaluatedItems(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	arr, ok := instance.([]any)
	if !ok {
		return true, nil
	}
	log, ok := ctx.Items.(itemLog)
	rec, ok2 := ctx.Items.(keywords.ItemRecorder)
	if !ok || !ok2 {
		// fail closed: without the record we cannot tell what is covered
		return false, errors.New("unevaluatedItems: evaluated-item record unavailable")
	}
	all, prefix, indices := log.CoveredItems(ctx.InstanceLocation)
	if all {
		return true, nil // nothing left to apply to; already recorded as covered
	}
	covered := make([]bool, len(arr))
	for i := 0; i <= prefix && i < len(covered); i++ {
		covered[i] = true
	}
	for _, i := range indices { // contains: matching indices only
		if i >= 0 && i < len(covered) {
			covered[i] = true
		}
	}
	valid, applied := true, 0
	for i, item := range arr {
		if covered[i] {
			continue
		}
		applied++
		if err := validation.CountStep(ctx); err != nil {
			return false, err
		}
		ok, units, err := ctx.EvalChild(compiled, item, ctx.InstanceLocation+"/"+strconv.Itoa(i))
		if err != nil {
			return false, err
		}
		if !ok {
			valid = false
			ctx.Keep(units)
		}
	}
	if valid && applied > 0 {
		// ponytail: "all items" is one compact entry, not one per index
		rec.RecordItems(ctx.InstanceLocation, ctx.KeywordLocation)
		return true, ctx.Annotate(true)
	}
	return valid, nil
}
