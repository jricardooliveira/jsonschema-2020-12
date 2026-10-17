package compile

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/neoface/jsonschema/internal/keywords"
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
	return CompileWithRegistry(doc, baseURI, nil)
}

// CompileWithRegistry is CompileDocument that also looks a reference up in
// ext (the supplied schemas) when the document itself has no such resource.
// Every supplied resource reached from the document is compiled and its own
// references resolved here, so nothing is left to resolve at Validate time.
// ext is only read. A nil ext means no supplied schemas.
func CompileWithRegistry(doc any, baseURI string, ext *registry.Index) (*Node, error) {
	_, root, err := compileDocument(doc, baseURI, ext)
	return root, err
}

// shared is the state common to the main document and the supplied ones.
type shared struct {
	main, ext *registry.Index
	byOwner   map[*registry.Owner]*resolver
	all       []*resolver
	pending   []pendingRef
	todo      []pendingRes // resources whose Dynamic table is not filled yet
	// formatAssertion: some resource's meta-schema requires format-assertion.
	formatAssertion bool
}

type pendingRef struct {
	r   *resolver
	ref *core.Ref
}

type pendingRes struct {
	r    *resolver
	res  *registry.Resource
	info *Resource
}

func (sh *shared) add(doc any, bases map[string]string, resources map[string]*registry.Resource) *resolver {
	r := &resolver{sh: sh, doc: doc, bases: bases, resources: resources, memo: map[string]*Node{}, infos: map[string]*Resource{}, vocabs: map[string]*VocabularySet{}}
	sh.all = append(sh.all, r)
	return r
}

// info returns the compiled resource holding schema location loc, creating
// it on first use. Its Dynamic table is filled later by fill, so compiling
// an anchor never re-enters a node that is still being compiled.
func (r *resolver) info(loc string) *Resource {
	base := r.baseAt(loc)
	if in, ok := r.infos[base]; ok {
		return in
	}
	res := r.resources[base]
	if res == nil {
		r.infos[base] = nil
		return nil
	}
	in := &Resource{URI: res.URI, Pointer: res.Pointer, Dynamic: map[string]*Node{}}
	r.infos[base] = in
	r.sh.todo = append(r.sh.todo, pendingRes{r, res, in})
	return in
}

// fill compiles every $dynamicAnchor subschema of p's resource into its table.
func (p pendingRes) fill() error {
	names := make([]string, 0, len(p.res.Anchors))
	for name := range p.res.Anchors {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if !p.res.HasDynamicAnchor(name) {
			continue
		}
		ptr := p.res.Anchors[name]
		v, ok := navigate(p.r.doc, splitPtr(ptr))
		if !ok {
			continue
		}
		n, err := p.r.compile(v, ptr)
		if err != nil {
			return err
		}
		p.info.Dynamic[name] = n
	}
	return nil
}

// CompileReached is CompileWithRegistry that also returns every document the
// main one reaches: the main schema first (Source ""), then each supplied
// document a reference led to. Built-in meta-schemas are not listed. On error
// the documents reached before the failure are still returned, so the caller
// can tell a meta-schema failure in one of them from other errors.
func CompileReached(doc any, baseURI string, ext *registry.Index) (*Node, []Doc, error) {
	r, root, err := compileDocument(doc, baseURI, ext)
	docs := []Doc{{Value: doc}}
	if r != nil {
		for _, o := range r.sh.all[1:] {
			if o.source != registry.BuiltinSource {
				docs = append(docs, Doc{Value: o.doc, Source: o.source})
			}
		}
	}
	if err != nil {
		return nil, docs, err
	}
	return root, docs, nil
}

func compileDocument(doc any, baseURI string, ext *registry.Index) (*resolver, *Node, error) {
	idx, err := registry.Build(doc, baseURI)
	if err != nil {
		return nil, nil, err
	}
	sh := &shared{main: idx, ext: ext, byOwner: map[*registry.Owner]*resolver{}}
	r := sh.add(doc, idx.Bases, idx.Resources)
	root, err := r.compile(doc, "")
	if err != nil {
		return r, nil, err
	}
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
				if _, err := r.compile(v, p); err != nil {
					return r, nil, err
				}
			}
		}
	}
	// Compiling a target appends the $refs it holds; every node is memoized,
	// so the order targets are reached never causes a recompile. Filling a
	// resource's dynamic-anchor table may add references and resources too.
	for i, j := 0, 0; i < len(sh.pending) || j < len(sh.todo); {
		if j < len(sh.todo) {
			if err := sh.todo[j].fill(); err != nil {
				return r, nil, err
			}
			j++
			continue
		}
		p := sh.pending[i]
		i++
		l, err := p.r.locate(p.ref)
		if err != nil {
			return r, nil, err
		}
		n, err := l.in.compile(l.v, l.ptr)
		if err != nil {
			return r, nil, fmt.Errorf("schema location #%s: $ref %q: %v", l.ref.Location, l.ref.Ref, err)
		}
		l.ref.Target = n
	}
	root.FormatAssertion = sh.formatAssertion
	return r, root, sh.checkPureCycles()
}

// located is a $ref with the document, schema pointer and value of its target.
type located struct {
	ref *core.Ref
	in  *resolver
	ptr string
	v   any
}

type resolver struct {
	sh        *shared
	doc       any
	source    string                        // where doc came from; "" for the main schema
	bases     map[string]string             // schema pointer -> base URI in effect there
	resources map[string]*registry.Resource // base URI -> indexed resource
	infos     map[string]*Resource          // base URI -> compiled resource (nil: not indexed)
	memo      map[string]*Node              // schema pointer -> compiled node
	vocabs    map[string]*VocabularySet     // base URI -> active set (nil: all)
	nodes     []*Node
	refs      []*core.Ref
}

// compile is compileAt with the memo consulted at every node it visits, so
// each schema location is compiled once, whatever order root and $ref
// targets reach it in. It records each new node and its $ref keywords.
// Locations are document pointers, so one location always holds one value.
// ponytail: mirrors compileAt's body; fold compileAt into this if compile.go
// is ever in scope.
func (r *resolver) compile(v any, loc string) (*Node, error) {
	if n, ok := r.memo[loc]; ok {
		return n, nil
	}
	s, isMap := v.(map[string]any)
	if !isMap {
		n, err := compileAt(v, loc) // boolean (or error): no children
		if err != nil {
			return nil, err
		}
		n.Resource = r.info(loc)
		r.memo[loc] = n
		r.nodes = append(r.nodes, n)
		return n, nil
	}
	names := make([]string, 0, len(s))
	for name := range s {
		names = append(names, name)
	}
	sort.Strings(names)
	n := &Node{Location: loc, Resource: r.info(loc)}
	vocab, err := r.vocabAt(loc)
	if err != nil {
		return nil, err
	}
	at := func(v any, at string) (any, error) {
		c, err := r.compile(v, at)
		if err != nil {
			return nil, err
		}
		return c, nil
	}
	for _, name := range names {
		kw, ok := keywords.Lookup(name)
		if !ok || !vocab.allows(kw) {
			// Unknown or inactive-vocabulary keywords are annotated (core 6.5, 8.1.2).
			n.Items = append(n.Items, Item{keywords.AnnotateOnly(name, ""), s[name], loc + "/" + escape.Replace(name)})
			continue
		}
		kwLoc := loc + "/" + escape.Replace(name)
		ctx := &keywords.CompileContext{
			Location:  kwLoc,
			Siblings:  s,
			Sub:       func(v any, rel string) (any, error) { return at(v, kwLoc+rel) },
			CompileAt: at,
		}
		data, err := kw.Compile(ctx, s[name])
		if err != nil {
			return nil, err
		}
		n.Items = append(n.Items, Item{kw, data, kwLoc})
		if ref, ok := data.(*core.Ref); ok {
			r.refs = append(r.refs, ref)
			r.sh.pending = append(r.sh.pending, pendingRef{r, ref})
		}
	}
	r.memo[loc] = n
	r.nodes = append(r.nodes, n)
	return n, nil
}

func (r *resolver) baseAt(ptr string) string {
	for {
		if b, ok := r.bases[ptr]; ok {
			return b
		}
		i := strings.LastIndexByte(ptr, '/')
		if i < 0 {
			return r.bases[""]
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
	in := r
	res := r.sh.main.Resources[key]
	if res == nil && r.sh.ext != nil {
		if res = r.sh.ext.Resources[key]; res != nil {
			if res.Owner == nil {
				return fail("unresolved reference, %q has no document", key)
			}
			if in = r.sh.byOwner[res.Owner]; in == nil {
				in = r.sh.add(res.Owner.Doc, res.Owner.Bases, r.sh.ext.Resources)
				in.source = res.Source
				r.sh.byOwner[res.Owner] = in
			}
		}
	}
	if res == nil {
		return fail("unresolved reference, %q is not a resource of the loaded schema or the supplied schemas", key)
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
	v, ok := navigate(in.doc, toks)
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
	return located{ref: ref, in: in, ptr: b.String(), v: v}, nil
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
func (sh *shared) checkPureCycles() error {
	clean := map[*Node]bool{}
	var nodes []*Node
	for _, r := range sh.all {
		nodes = append(nodes, r.nodes...)
	}
	for _, start := range nodes {
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
