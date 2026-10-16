// Package registry indexes the schema resources and anchors of one loaded
// schema document. It never reads disk or network.
package registry

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/neoface/jsonschema/internal/jsonvalue"
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
	Doc     any               // the document holding the resource
	Source  string            // where Doc came from; empty for the main schema
	Owner   *Owner            // set on supplied resources: the document they live in
}

// Owner is one supplied document and the base URI of each of its schemas,
// shared by every resource found in that document.
type Owner struct {
	Doc   any
	Bases map[string]string // schema JSON Pointer -> base URI in effect there
}

// value returns the schema the resource points at.
func (r *Resource) value() any {
	v, _ := navigate(r.Doc, r.Pointer)
	return v
}

func (r *Resource) from() string {
	if r.Source == "" {
		return "the main schema"
	}
	return r.Source
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

type pendingRef struct{ base, ref string }

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
	var refs []pendingRef
	resAt := map[string]*Resource{}
	queued := map[string]bool{}
	done := 0
	for {
		for len(stack) > 0 {
			it := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			m, ok := it.node.(map[string]any)
			if !ok { // boolean schemas (or non-schemas) carry no keywords
				if it.resource == nil && first {
					first = false
					if err := idx.add(&Resource{URI: it.base, Pointer: it.ptr, Doc: doc}, it.ptr); err != nil {
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
				res = &Resource{URI: base, Pointer: it.ptr, Doc: doc}
				if err := idx.add(res, it.ptr); err != nil {
					return nil, err
				}
			}
			first = false
			idx.Bases[it.ptr] = base
			resAt[it.ptr] = res
			if r, ok := m["$ref"].(string); ok {
				refs = append(refs, pendingRef{base, r})
			}
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
		// Follow $ref targets the keyword walk did not reach (for example under
		// unknown keywords), so an $id inside them starts its own resource.
		for ; done < len(refs); done++ {
			pr := refs[done]
			full, err := uri.Resolve(pr.base, pr.ref)
			if err != nil {
				continue // compile reports it
			}
			docPart, frag, _ := strings.Cut(full, "#")
			key, err := uri.Normalize(docPart)
			res := idx.Resources[key]
			if err != nil || res == nil || !strings.HasPrefix(frag, "/") {
				continue
			}
			toks, err := uri.ParsePointer(frag)
			if err != nil {
				continue
			}
			ptr := res.Pointer
			for _, t := range toks {
				ptr += "/" + Escape(t)
			}
			if _, seen := idx.Bases[ptr]; seen || queued[ptr] {
				continue
			}
			v, ok := navigate(doc, ptr)
			if !ok {
				continue
			}
			switch v.(type) {
			case map[string]any, bool:
			default:
				continue
			}
			anc := ptr
			for resAt[anc] == nil {
				i := strings.LastIndexByte(anc, '/')
				if i < 0 {
					break
				}
				anc = anc[:i]
			}
			if resAt[anc] == nil {
				continue
			}
			queued[ptr] = true
			stack = append(stack, item{node: v, ptr: ptr, base: idx.Bases[anc], resource: resAt[anc]})
		}
		if len(stack) == 0 {
			break
		}
	}
	return idx, nil
}

// navigate follows an escaped JSON Pointer in v.
func navigate(v any, ptr string) (any, bool) {
	if ptr == "" {
		return v, true
	}
	toks, err := uri.ParsePointer(ptr)
	if err != nil {
		return nil, false
	}
	for _, t := range toks {
		switch c := v.(type) {
		case map[string]any:
			var ok bool
			if v, ok = c[t]; !ok {
				return nil, false
			}
		case []any:
			i, err := strconv.Atoi(t)
			if err != nil || i < 0 || i >= len(c) || strconv.Itoa(i) != t {
				return nil, false
			}
			v = c[i]
		default:
			return nil, false
		}
	}
	return v, true
}

// AddExternal indexes a supplied document under its address (when not empty),
// its $id and its embedded $ids. source names the file in errors. Two
// documents claiming one address are accepted only when deep-equal.
func (idx *Index) AddExternal(doc any, address, source string) error {
	sub, err := Build(doc, address)
	if err != nil {
		return fmt.Errorf("%s: %v", source, err)
	}
	var root *Resource
	owner := &Owner{Doc: doc, Bases: sub.Bases}
	keys := make([]string, 0, len(sub.Resources)+1)
	for k, r := range sub.Resources {
		r.Source = source
		r.Owner = owner
		keys = append(keys, k)
		if r.Pointer == "" {
			root = r
		}
	}
	if address != "" && root != nil {
		a, err := stripFrag(address)
		if err != nil {
			return fmt.Errorf("%s: %v", source, err)
		}
		if _, ok := sub.Resources[a]; !ok { // root has its own $id: both addresses reach it
			sub.Resources[a] = root
			keys = append(keys, a)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !strings.Contains(k, ":") { // not absolute: nothing can reach it
			continue
		}
		r := sub.Resources[k]
		if prev, dup := idx.Resources[k]; dup {
			if !jsonvalue.Equal(prev.value(), r.value()) {
				return fmt.Errorf("duplicate address %q claimed by %s and %s", k, prev.from(), r.from())
			}
			continue
		}
		idx.Resources[k] = r
	}
	return nil
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
