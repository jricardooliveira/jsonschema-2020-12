package unevaluated

import (
	"errors"
	"fmt"
	"sort"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords/validation"
)

// VocabularyURI is the 2020-12 unevaluated vocabulary.
const VocabularyURI = "https://json-schema.org/draft/2020-12/vocab/unevaluated"

func init() {
	keywords.Register(keywords.Keyword{
		Name: "unevaluatedProperties", Vocabulary: VocabularyURI,
		Compile: compileUnevaluatedProperties, Evaluate: evalUnevaluatedProperties,
		AfterSiblings: true, // core 11.3: sees every sibling's annotations
	})
}

// propLog is the per-call record of evaluated property names (an
// *eval.EvaluatedItems); failed branches are already rolled back from it.
type propLog interface {
	RecordProperties(instLoc, kwLoc string, names []string)
	CoveredProperties(mark int, instLoc string) (map[string]bool, error)
}

func compileUnevaluatedProperties(ctx *keywords.CompileContext, value any) (any, error) {
	n, err := ctx.Sub(value, "")
	if err != nil {
		return nil, fmt.Errorf("unevaluatedProperties: %w", err)
	}
	return n, nil
}

// evalUnevaluatedProperties applies the subschema to every member that no
// successful evaluation at this instance location covered (core 11.3). Each
// failing member yields its own units at its own instance location.
func evalUnevaluatedProperties(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	obj, ok := instance.(map[string]any)
	if !ok {
		return true, nil
	}
	log, ok := ctx.Items.(propLog)
	if !ok {
		// fail closed: without the record we cannot tell what is covered
		return false, errors.New("unevaluatedProperties: evaluated-property record unavailable")
	}
	// only this schema object's own record counts (core 11.3)
	covered, err := log.CoveredProperties(ctx.Scope, ctx.InstanceLocation)
	if err != nil {
		return false, err
	}
	var applied []string
	for k := range obj {
		if !covered[k] {
			applied = append(applied, k)
		}
	}
	sort.Strings(applied)
	valid := true
	for _, k := range applied {
		if err := validation.CountStep(ctx); err != nil {
			return false, err
		}
		loc := ctx.InstanceLocation + "/" + pointerEscape(k)
		ok, units, err := ctx.EvalChild(compiled, obj[k], loc)
		if err != nil {
			return false, err
		}
		if !ok {
			valid = false
			ctx.Keep(units)
		}
	}
	log.RecordProperties(ctx.InstanceLocation, ctx.KeywordLoc(), applied)
	if valid && len(applied) > 0 {
		return true, ctx.Annotate(append([]string(nil), applied...))
	}
	return valid, nil
}

func pointerEscape(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '~':
			out = append(out, '~', '0')
		case '/':
			out = append(out, '~', '1')
		default:
			out = append(out, s[i])
		}
	}
	return string(out)
}
