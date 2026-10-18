package applicator

import (
	"fmt"
	"strconv"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

// compileSchemaArray compiles a non-empty array of schemas, element i at
// /<keyword>/<i>. Any other value is a schema error naming the keyword.
func compileSchemaArray(name string, ctx *keywords.CompileContext, value any) ([]any, error) {
	list, ok := value.([]any)
	if !ok || len(list) == 0 {
		return nil, fmt.Errorf("%s: must be a non-empty array of schemas", name)
	}
	nodes := make([]any, len(list))
	for i, v := range list {
		n, err := ctx.Sub(v, "/"+strconv.Itoa(i))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		nodes[i] = n
	}
	return nodes, nil
}
