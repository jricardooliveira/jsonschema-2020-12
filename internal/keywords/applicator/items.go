package applicator

import (
	"fmt"

	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "items", Vocabulary: VocabularyURI, Compile: compileItems, Evaluate: evalItems})
}

// compiledItems is built once at compile time and only read afterwards.
type compiledItems struct {
	node   any
	prefix int // length of the sibling prefixItems
}

func compileItems(ctx *keywords.CompileContext, value any) (any, error) {
	n, err := ctx.Sub(value, "")
	if err != nil {
		return nil, fmt.Errorf("items: %w", err)
	}
	prefix := 0
	if p, ok := ctx.Siblings["prefixItems"].([]any); ok {
		prefix = len(p)
	}
	return &compiledItems{n, prefix}, nil
}

func evalItems(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	arr, ok := instance.([]any)
	if !ok {
		return true, nil
	}
	c := compiled.(*compiledItems)
	if len(arr) <= c.prefix {
		return true, nil
	}
	valid, err := evalElements(ctx, c.node, arr, c.prefix, len(arr))
	if err != nil {
		return false, err
	}
	if valid { // core 10.3.1.2: annotation true when items applied
		if rec, ok := ctx.Items.(keywords.ItemRecorder); ok {
			rec.RecordItems(ctx.InstanceLocation, ctx.KeywordLocation)
		}
		return true, ctx.Annotate(true)
	}
	return valid, nil
}
