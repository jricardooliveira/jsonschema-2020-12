package applicator

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/neoface/jsonschema/internal/keywords"
)

// VocabularyURI is the 2020-12 applicator vocabulary.
const VocabularyURI = "https://json-schema.org/draft/2020-12/vocab/applicator"

func init() {
	keywords.Register(keywords.Keyword{Name: "prefixItems", Vocabulary: VocabularyURI, Compile: compilePrefixItems, Evaluate: evalPrefixItems})
}

func compilePrefixItems(ctx *keywords.CompileContext, value any) (any, error) {
	list, ok := value.([]any)
	if !ok {
		return nil, errors.New("prefixItems: must be an array of schemas")
	}
	nodes := make([]any, len(list))
	for i, v := range list {
		n, err := ctx.Sub(v, "/"+strconv.Itoa(i))
		if err != nil {
			return nil, fmt.Errorf("prefixItems: %w", err)
		}
		nodes[i] = n
	}
	return nodes, nil
}

// evalElements checks arr[from:to] against node and keeps every failing
// element's units. It never stops at the first failure.
func evalElements(ctx *keywords.EvalContext, node any, arr []any, from, to int) (bool, error) {
	valid := true
	for i := from; i < to; i++ {
		ok, units, err := ctx.EvalChild(node, arr[i], ctx.InstanceLocation+"/"+strconv.Itoa(i))
		if err != nil {
			return false, err
		}
		if !ok {
			valid = false
			ctx.Keep(units)
		}
	}
	return valid, nil
}

func evalPrefixItems(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	arr, ok := instance.([]any)
	if !ok {
		return true, nil
	}
	nodes := compiled.([]any)
	n := min(len(arr), len(nodes))
	valid := true
	for i := 0; i < n; i++ {
		ok, err := evalElements(ctx, nodes[i], arr, i, i+1)
		if err != nil {
			return false, err
		}
		valid = valid && ok
	}
	// Annotations (core 10.3.1.1) only for a keyword that passed and applied.
	if valid && n > 0 {
		if rec, ok := ctx.Items.(keywords.ItemRecorder); ok {
			rec.RecordPrefixItems(ctx.InstanceLocation, ctx.KeywordLoc(), n-1, n == len(arr))
		}
		var v any = n - 1
		if n == len(arr) {
			v = true
		}
		return true, ctx.Annotate(v)
	}
	return valid, nil
}
