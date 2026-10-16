// Package registry indexes the schema resources and anchors of one loaded
// schema document. It never reads disk or network.
package registry

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/neoface/jsonschema/internal/uri"
)

var anchorRe = regexp.MustCompile(`^[A-Za-z_][-A-Za-z0-9._]*$`)

// Subschema-bearing keywords of 2020-12 (core 10, validation, applicators).
var (
	singleKw = []string{"additionalProperties", "items", "contains", "not", "if", "then", "else",
		"propertyNames", "unevaluatedItems", "unevaluatedProperties", "contentSchema"}
	arrayKw = []string{"allOf", "anyOf", "oneOf", "prefixItems"}
	mapKw   = []string{"$defs", "properties", "patternProperties", "dependentSchemas"}
)

// Resource is one schema resource: the root document or an embedded $id.
type Resource struct {
	URI     string            // normalized absolute URI, no fragment
	Pointer string            // JSON Pointer of the resource root in the document
	Anchors map[string]string // anchor name -> JSON Pointer of the schema
}

// Index is the result of Build. It is read-only once built.
type Index struct {
	Resources map[string]*Resource // by normalized URI
	Bases     map[string]string    // schema JSON Pointer -> base URI in effect there
}

// Escape escapes one JSON Pointer token (~ then /).
func Escape(tok string) string {
	return strings.ReplaceAll(strings.ReplaceAll(tok, "~", "~0"), "/", "~1")
}

type item struct {
	node     any
	ptr      string
	base     string
	resource *Resource
}

// Build walks doc (decoded JSON) and indexes its resources. rootBase is the
// base URI used when the root has no $id. Errors name the schema location.
func Build(doc any, rootBase string) (*Index, error) {
	idx := &Index{Resources: map[string]*Resource{}, Bases: map[string]string{}}
	base, err := stripFrag(rootBase)
	if err != nil {
		return nil, fmt.Errorf("base URI %q: %v", rootBase, err)
	}
	stack := []item{{node: doc, ptr: "", base: base}}
	first := true
	for len(stack) > 0 {
		it := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		m, ok := it.node.(map[string]any)
		if !ok { // boolean schemas (or non-schemas) carry no keywords
			if it.resource == nil && first {
				first = false
				if err := idx.add(&Resource{URI: it.base, Pointer: it.ptr}, it.ptr); err != nil {
					return nil, err
				}
			}
			idx.Bases[it.ptr] = it.base
			continue
		}
		res := it.resource
		base := it.base
		if id, ok := m["$id"].(string); ok {
			full, err := uri.Resolve(it.base, id)
			if err != nil {
				return nil, fmt.Errorf("%s: $id %q: %v", loc(it.ptr), id, err)
			}
			if i := strings.IndexByte(full, '#'); i >= 0 && i < len(full)-1 {
				return nil, fmt.Errorf("%s: $id %q must not have a non-empty fragment", loc(it.ptr), id)
			}
			if base, err = stripFrag(full); err != nil {
				return nil, fmt.Errorf("%s: $id %q: %v", loc(it.ptr), id, err)
			}
			res = nil
		}
		if res == nil {
			res = &Resource{URI: base, Pointer: it.ptr}
			if err := idx.add(res, it.ptr); err != nil {
				return nil, err
			}
		}
		first = false
		idx.Bases[it.ptr] = base
		for _, kw := range []string{"$anchor", "$dynamicAnchor"} {
			name, ok := m[kw].(string)
			if !ok {
				continue
			}
			if !anchorRe.MatchString(name) {
				return nil, fmt.Errorf("%s: %s %q is not a valid anchor name", loc(it.ptr), kw, name)
			}
			if prev, dup := res.Anchors[name]; dup && prev != it.ptr {
				return nil, fmt.Errorf("%s: duplicate anchor %q in resource %q", loc(it.ptr), name, res.URI)
			}
			if res.Anchors == nil {
				res.Anchors = map[string]string{}
			}
			res.Anchors[name] = it.ptr
		}
		push := func(child any, p string) {
			stack = append(stack, item{node: child, ptr: p, base: base, resource: res})
		}
		for _, kw := range singleKw {
			if c, ok := m[kw]; ok {
				push(c, it.ptr+"/"+kw)
			}
		}
		for _, kw := range arrayKw {
			if a, ok := m[kw].([]any); ok {
				for i, c := range a {
					push(c, fmt.Sprintf("%s/%s/%d", it.ptr, kw, i))
				}
			}
		}
		for _, kw := range mapKw {
			if sub, ok := m[kw].(map[string]any); ok {
				keys := make([]string, 0, len(sub))
				for k := range sub {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					push(sub[k], it.ptr+"/"+kw+"/"+Escape(k))
				}
			}
		}
	}
	return idx, nil
}

func (idx *Index) add(r *Resource, ptr string) error {
	if _, dup := idx.Resources[r.URI]; dup {
		return fmt.Errorf("%s: duplicate $id %q", loc(ptr), r.URI)
	}
	idx.Resources[r.URI] = r
	return nil
}

// stripFrag normalizes s and drops an empty fragment; a non-empty one is an error.
func stripFrag(s string) (string, error) {
	n, err := uri.Normalize(s)
	if err != nil {
		return "", err
	}
	if i := strings.IndexByte(n, '#'); i >= 0 {
		if i < len(n)-1 {
			return "", fmt.Errorf("non-empty fragment")
		}
		n = n[:i]
	}
	return n, nil
}

func loc(ptr string) string { return "schema location #" + ptr }
