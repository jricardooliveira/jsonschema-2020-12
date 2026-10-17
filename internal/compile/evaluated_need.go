package compile

import (
	"strings"

	"github.com/neoface/jsonschema/internal/registry"
)

// NeedsEvaluated reports whether unevaluatedProperties or unevaluatedItems
// could be reached when evaluating the main schema (docs[0]): it answers true
// unless it can prove neither keyword appears anywhere evaluation could go.
//
// ponytail: cautious by construction. It scans every document of docs (the
// main schema and every supplied schema the compiler reached) for the keyword
// names anywhere in the value, with no per-subschema tracking, so cycles of
// $ref cannot matter. When any document uses a reference that leaves the
// document (a $ref or $dynamicRef not starting with '#'), a $dynamicRef could
// land on any registry resource, so every resource in ext (supplied and
// embedded meta-schemas) is scanned too. Upgrade to a real reachability walk
// only if this cautious rule costs real workloads the saving.
func NeedsEvaluated(docs []Doc, ext *registry.Index) bool {
	external := false
	for _, d := range docs {
		found, ext := scan(d.Value)
		if found {
			return true
		}
		external = external || ext
	}
	if !external {
		return false
	}
	if ext == nil {
		return true // cannot enumerate the registry: stay cautious
	}
	for _, r := range ext.Resources {
		if found, _ := scan(r.Doc); found {
			return true
		}
	}
	return false
}

// scan walks v with an explicit stack. found: an unevaluated* keyword name
// occurs; external: a reference leaves its document.
func scan(v any) (found, external bool) {
	stack := []any{v}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		switch x := cur.(type) {
		case map[string]any:
			for k, val := range x {
				if k == "unevaluatedProperties" || k == "unevaluatedItems" {
					return true, external
				}
				if s, ok := val.(string); ok && (k == "$ref" || k == "$dynamicRef") && !strings.HasPrefix(s, "#") {
					external = true
				}
				stack = append(stack, val)
			}
		case []any:
			stack = append(stack, x...)
		}
	}
	return false, external
}
