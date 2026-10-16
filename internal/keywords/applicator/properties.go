// Ordering rule for the object keywords (operator decision): keywords are
// evaluated and reported in sorted keyword-name order; within one keyword,
// member failures come in sorted member-name order (sort.Strings, byte-wise
// UTF-8); dependentSchemas and dependentRequired walk their trigger names
// sorted; missing required/dependentRequired names follow the schema array
// order. A consequence: dependentRequired failures come before properties
// failures in the same object, because "dependentRequired" sorts first.

package applicator

import (
	"fmt"
	"sort"

	"github.com/neoface/jsonschema/internal/keywords"
	"github.com/neoface/jsonschema/internal/keywords/validation"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "properties", Vocabulary: VocabularyURI, Compile: compileProperties, Evaluate: evalProperties})
}

// propRecorder is what the three property keywords need from ctx.Items.
type propRecorder interface {
	RecordProperties(instLoc, kwLoc string, names []string)
}

// recordProps records the applied names for unevaluated* tracking and, when
// the keyword passed, annotates them (core 10.3.2). Names are already in the
// instance location; no instance value is annotated.
func recordProps(ctx *keywords.EvalContext, names []string, valid bool) error {
	if rec, ok := ctx.Items.(propRecorder); ok {
		rec.RecordProperties(ctx.InstanceLocation, ctx.KeywordLocation, names)
	}
	if valid && len(names) > 0 {
		return ctx.Annotate(append([]string(nil), names...))
	}
	return nil
}

// step counts one evaluation step (see validation.CountStep).
var step = validation.CountStep

// sortedKeys returns a sorted copy of the object's member names.
func sortedKeys(obj map[string]any) []string {
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// evalMember applies node to the member key; failures are kept. The key
// appears only in the instance location.
func evalMember(ctx *keywords.EvalContext, node any, obj map[string]any, key string) (bool, error) {
	ok, units, err := ctx.EvalChild(node, obj[key], ctx.InstanceLocation+"/"+depPointerEscape.Replace(key))
	if err != nil {
		return false, err
	}
	if !ok {
		ctx.Keep(units)
		if err := step(ctx); err != nil {
			return false, err
		}
	}
	return ok, nil
}

func compileProperties(ctx *keywords.CompileContext, value any) (any, error) {
	m, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("properties: must be an object")
	}
	nodes := make(map[string]any, len(m))
	for _, k := range sortedKeys(m) {
		n, err := ctx.Sub(m[k], "/"+depPointerEscape.Replace(k))
		if err != nil {
			return nil, fmt.Errorf("properties: %w", err)
		}
		nodes[k] = n
	}
	return nodes, nil
}

func evalProperties(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	obj, ok := instance.(map[string]any)
	if !ok {
		return true, nil
	}
	nodes := compiled.(map[string]any)
	valid := true
	var applied []string
	for _, k := range sortedKeys(obj) {
		if err := step(ctx); err != nil {
			return false, err
		}
		n, has := nodes[k]
		if !has {
			continue
		}
		applied = append(applied, k)
		ok, err := evalMember(ctx, n, obj, k)
		if err != nil {
			return false, err
		}
		valid = valid && ok
	}
	return valid, recordProps(ctx, applied, valid)
}
