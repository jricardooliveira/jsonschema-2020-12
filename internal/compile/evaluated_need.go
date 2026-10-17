package compile

import (
	"strings"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/registry"
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
		found, e := scan(d.Value)
		if found {
			return true
		}
		external = external || e
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

// scan walks v. found: an unevaluated* keyword name occurs; external: a
// reference leaves its document.
// ponytail: plain recursion, so the scan allocates nothing at Compile. Its
// depth is bounded because every schema value was decoded under MaxDepth
// (main schema and registry files) or is an embedded meta-schema.
func scan(v any) (found, external bool) {
	switch x := v.(type) {
	case map[string]any:
		for k, val := range x {
			if k == "unevaluatedProperties" || k == "unevaluatedItems" {
				return true, external
			}
			if s, ok := val.(string); ok && (k == "$ref" || k == "$dynamicRef") && !strings.HasPrefix(s, "#") {
				external = true
			}
			f, e := scan(val)
			if f {
				return true, external
			}
			external = external || e
		}
	case []any:
		for _, val := range x {
			f, e := scan(val)
			if f {
				return true, external
			}
			external = external || e
		}
	}
	return false, external
}
