package compile

import (
	"fmt"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/neoface/jsonschema/internal/keywords/core"
	"github.com/neoface/jsonschema/internal/registry"
	"github.com/neoface/jsonschema/internal/uri"
)

// StdinBaseURI is the fixed base URI of a schema that has no file and no
// override. The .invalid TLD (RFC 2606) never resolves, and nothing is read.
const StdinBaseURI = "https://stdin.invalid/schema.json"

var schemeRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*:`)

// RootBaseURI picks the base URI of a root schema without $id: the override
// when given, else the file location of filePath, else StdinBaseURI. It is
// pure string work: no file is stat'ed, opened or resolved.
func RootBaseURI(override, filePath string) (string, error) {
	if override != "" {
		if !schemeRe.MatchString(override) {
			return "", fmt.Errorf("base URI %q must be absolute", override)
		}
		n, err := uri.Normalize(override)
		if err != nil {
			return "", fmt.Errorf("base URI %q: %v", override, err)
		}
		if i := strings.IndexByte(n, '#'); i >= 0 {
			if i < len(n)-1 {
				return "", fmt.Errorf("base URI %q must not have a fragment", override)
			}
			n = n[:i]
		}
		return n, nil
	}
	if filePath == "" {
		return StdinBaseURI, nil
	}
	p, err := filepath.Abs(filePath) // may read the working directory, never the path
	if err != nil {
		return "", fmt.Errorf("base URI from %q: %v", filePath, err)
	}
	return (&url.URL{Scheme: "file", Path: filepath.ToSlash(p)}).String(), nil
}

// CompileDocument compiles doc and resolves every $ref once against the
// resource index of doc, so a broken reference fails here whether or not any
// document would reach it. baseURI is the base of a root without $id. Targets
// that are objects or booleans are followed even under unknown keywords
// (core 9.4.2 leaves this undefined); any other value is an error.
func CompileDocument(doc any, baseURI string) (*Node, error) {
	_, root, err := compileDocument(doc, baseURI)
	return root, err
}

func compileDocument(doc any, baseURI string) (*resolver, *Node, error) {
	idx, err := registry.Build(doc, baseURI)
	if err != nil {
		return nil, nil, err
	}
	root, err := compileAt(doc, "")
	if err != nil {
		return nil, nil, err
	}
	r := &resolver{doc: doc, idx: idx, memo: map[string]*Node{"": root}}
	r.add(root)
	// References inside $defs and other uncompiled branches are checked too.
	ptrs := make([]string, 0, len(idx.Bases))
	for p := range idx.Bases {
		ptrs = append(ptrs, p)
	}
	sort.Strings(ptrs)
	for _, p := range ptrs {
		v, ok := navigate(doc, splitPtr(p))
		if m, isMap := v.(map[string]any); ok && isMap {
			if _, has := m["$ref"]; has {
				if _, err := r.target(p, v); err != nil {
					return nil, nil, err
				}
			}
		}
	}
	// Resolve in batches; each batch compiles its targets in pointer order so
	// an ancestor is compiled (and memoizes its subtree) before a descendant.
	// ponytail: a later batch can still recompile below an earlier target
	// under an unknown keyword; add a node budget if that ever matters.
	for start := 0; start < len(r.refs); {
		end := len(r.refs)
		batch := make([]located, 0, end-start)
		for _, ref := range r.refs[start:end] {
			l, err := r.locate(ref)
			if err != nil {
				return nil, nil, err
			}
			batch = append(batch, l)
		}
		sort.SliceStable(batch, func(i, j int) bool { return batch[i].ptr < batch[j].ptr })
		for _, l := range batch {
			n, err := r.target(l.ptr, l.v)
			if err != nil {
				return nil, nil, fmt.Errorf("schema location #%s: $ref %q: %v", l.ref.Location, l.ref.Ref, err)
			}
			l.ref.Target = n
		}
		start = end
	}
	return r, root, r.checkPureCycles()
}

// located is a $ref with the schema pointer and value of its target.
type located struct {
	ref *core.Ref
	ptr string
	v   any
}

type resolver struct {
	doc   any
	idx   *registry.Index
	memo  map[string]*Node // schema pointer -> compiled node
	nodes []*Node
	refs  []*core.Ref
}

// add records n and every node reachable from it, and their $ref keywords.
// Each node is memoized by location, so no later target compiles it again.
func (r *resolver) add(n *Node) {
	for _, m := range collectNodes(n) {
		if _, ok := r.memo[m.Location]; !ok {
			r.memo[m.Location] = m
		}
		r.nodes = append(r.nodes, m)
		for _, it := range m.Items {
			if ref, ok := it.Data.(*core.Ref); ok {
				r.refs = append(r.refs, ref)
			}
		}
	}
}

func (r *resolver) target(ptr string, v any) (*Node, error) {
	if n, ok := r.memo[ptr]; ok {
		return n, nil
	}
	n, err := compileAt(v, ptr)
	if err != nil {
		return nil, err
	}
	r.memo[ptr] = n
	r.add(n)
	return n, nil
}

func (r *resolver) baseAt(ptr string) string {
	for {
		if b, ok := r.idx.Bases[ptr]; ok {
			return b
		}
		i := strings.LastIndexByte(ptr, '/')
		if i < 0 {
			return r.idx.Bases[""]
		}
		ptr = ptr[:i]
	}
}

// locate finds the schema pointer and value a $ref points at.
func (r *resolver) locate(ref *core.Ref) (located, error) {
	where := fmt.Sprintf("schema location #%s: $ref %q", ref.Location, ref.Ref)
	fail := func(format string, a ...any) (located, error) {
		return located{}, fmt.Errorf("%s: "+format, append([]any{where}, a...)...)
	}
	full, err := uri.Resolve(r.baseAt(strings.TrimSuffix(ref.Location, "/$ref")), ref.Ref)
	if err != nil {
		return fail("%v", err)
	}
	docPart, frag, _ := strings.Cut(full, "#")
	key, err := uri.Normalize(docPart)
	if err != nil {
		return fail("%v", err)
	}
	res := r.idx.Resources[key]
	if res == nil {
		return fail("unresolved reference, %q is not a resource of the loaded schema", key)
	}
	toks := splitPtr(res.Pointer)
	switch {
	case frag == "":
	case frag[0] == '/':
		more, err := uri.ParsePointer(frag)
		if err != nil {
			return fail("%v", err)
		}
		toks = append(toks, more...)
	default:
		name, err := url.PathUnescape(frag) // validated by uri.Resolve; '+' stays literal
		if err != nil {
			return fail("%v", err)
		}
		p, ok := res.Anchors[name]
		if !ok {
			return fail("unresolved reference, no anchor %q in %q", name, key)
		}
		toks = splitPtr(p)
	}
	v, ok := navigate(r.doc, toks)
	if !ok {
		return fail("unresolved reference, target does not exist")
	}
	switch v.(type) {
	case map[string]any, bool:
	default:
		return fail("unresolved reference, target is not an object or boolean schema")
	}
	var b strings.Builder
	for _, t := range toks {
		b.WriteString("/" + registry.Escape(t))
	}
	return located{ref: ref, ptr: b.String(), v: v}, nil
}

// pureRef returns n's $ref when n holds nothing else that can affect a result.
func pureRef(n *Node) *core.Ref {
	var ref *core.Ref
	for _, it := range n.Items {
		switch it.Keyword.Name {
		case "$ref":
			ref, _ = it.Data.(*core.Ref)
		case "$id", "$anchor", "$defs", "$comment":
		default:
			return nil
		}
	}
	return ref
}

// checkPureCycles rejects chains of schemas holding only $ref that loop.
func (r *resolver) checkPureCycles() error {
	clean := map[*Node]bool{}
	for _, start := range r.nodes {
		var path []*Node
		seen := map[*Node]bool{}
		for n := start; n != nil && !clean[n]; {
			ref := pureRef(n)
			if ref == nil {
				break
			}
			if seen[n] {
				return fmt.Errorf("schema location #%s: $ref %q is part of a cycle of references with no other keyword", ref.Location, ref.Ref)
			}
			seen[n] = true
			path = append(path, n)
			n, _ = ref.Target.(*Node)
		}
		for _, n := range path {
			clean[n] = true
		}
	}
	return nil
}

func splitPtr(p string) []string {
	if p == "" {
		return nil
	}
	toks := strings.Split(p[1:], "/")
	for i, t := range toks {
		toks[i] = strings.ReplaceAll(strings.ReplaceAll(t, "~1", "/"), "~0", "~")
	}
	return toks
}

func navigate(v any, toks []string) (any, bool) {
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

const modulePath = "github.com/neoface/jsonschema"

var nodeType = reflect.TypeOf((*Node)(nil))

// collectNodes finds every *Node reachable from n through keyword data.
// Keyword data is opaque to this package, so it is walked by reflection,
// read-only, and only through types of this module.
func collectNodes(n *Node) []*Node {
	var out []*Node
	seen := map[uintptr]bool{}
	var walk func(v reflect.Value)
	walk = func(v reflect.Value) {
		switch v.Kind() {
		case reflect.Interface:
			walk(v.Elem())
		case reflect.Pointer:
			if v.IsNil() || !strings.HasPrefix(v.Type().Elem().PkgPath(), modulePath) || seen[v.Pointer()] {
				return
			}
			seen[v.Pointer()] = true
			if v.Type() == nodeType { // UnsafePointer: Interface() panics on unexported fields
				out = append(out, (*Node)(v.UnsafePointer()))
			}
			walk(v.Elem())
		case reflect.Struct:
			if !strings.HasPrefix(v.Type().PkgPath(), modulePath) {
				return
			}
			for i := 0; i < v.NumField(); i++ {
				walk(v.Field(i))
			}
		case reflect.Slice, reflect.Array:
			switch v.Type().Elem().Kind() {
			case reflect.Interface, reflect.Pointer, reflect.Struct, reflect.Slice, reflect.Map, reflect.Array:
				for i := 0; i < v.Len(); i++ {
					walk(v.Index(i))
				}
			}
		case reflect.Map:
			for it := v.MapRange(); it.Next(); {
				walk(it.Value())
			}
		}
	}
	walk(reflect.ValueOf(n))
	return out
}
