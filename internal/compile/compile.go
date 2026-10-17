// Package compile turns a decoded schema into an immutable tree of nodes,
// finding keywords only through the keyword registry.
package compile

import (
	"errors"
	"sort"
	"strings"

	"github.com/neoface/jsonschema/internal/keywords"
)

// ErrNotSchema is returned for a schema that is neither a boolean nor an object.
var ErrNotSchema = errors.New("schema must be a boolean or an object")

// Item is one compiled keyword.
type Item struct {
	Keyword keywords.Keyword
	Data    any
	// Location is the JSON Pointer of the keyword from the schema root.
	Location string
}

// Node is a compiled schema. It is never modified after Compile returns.
type Node struct {
	Bool *bool // set for boolean schemas
	// Location is the JSON Pointer of this schema from the schema root.
	Location string
	Items    []Item
	// Resource is the schema resource holding this node; nil when compiled
	// without a resource index (Compile).
	Resource *Resource
	// FormatAssertion is set on the root node only: some resource's
	// meta-schema requires the format-assertion vocabulary.
	FormatAssertion bool
	// NoUnevaluated is set on the root node only, by NeedsEvaluated at load
	// time: no unevaluated* keyword is reachable. Zero (false) means unknown.
	NoUnevaluated bool
}

// Resource is the compiled view of one schema resource, shared by its nodes.
type Resource struct {
	URI     string // absolute URI, no fragment
	Pointer string // JSON Pointer of the resource root in its document
	// Dynamic maps each $dynamicAnchor name of the resource to its compiled
	// subschema (core 8.2.2), filled before Compile returns.
	Dynamic map[string]*Node
}

// AbsoluteLocation is the absolute keyword location (core 12.3.2) of the
// schema location loc inside resource r, or "" when r is nil.
func (r *Resource) AbsoluteLocation(loc string) string {
	if r == nil {
		return ""
	}
	return r.URI + "#" + encodeFragment(strings.TrimPrefix(loc, r.Pointer))
}

// Compile compiles a decoded schema value. Unknown keywords are ignored.
func Compile(v any) (*Node, error) { return compileAt(v, "") }

var escape = strings.NewReplacer("~", "~0", "/", "~1")

func compileAt(v any, loc string) (*Node, error) {
	switch s := v.(type) {
	case bool:
		return &Node{Bool: &s, Location: loc}, nil
	case map[string]any:
		names := make([]string, 0, len(s))
		for n := range s {
			names = append(names, n)
		}
		sort.Strings(names)
		n := &Node{Location: loc}
		for _, name := range names {
			kw, ok := keywords.Lookup(name)
			if !ok {
				// Unknown keywords never affect validity; they are annotated (core 6.5).
				n.Items = append(n.Items, Item{keywords.AnnotateOnly(name, ""), s[name], loc + "/" + escape.Replace(name)})
				continue
			}
			kwLoc := loc + "/" + escape.Replace(name)
			ctx := &keywords.CompileContext{
				Location: kwLoc,
				Siblings: s,
				Sub: func(v any, rel string) (any, error) {
					c, err := compileAt(v, kwLoc+rel)
					if err != nil {
						return nil, err
					}
					return c, nil
				},
				CompileAt: func(v any, at string) (any, error) {
					c, err := compileAt(v, at)
					if err != nil {
						return nil, err
					}
					return c, nil
				},
			}
			data, err := kw.Compile(ctx, s[name])
			if err != nil {
				return nil, err
			}
			n.Items = append(n.Items, Item{kw, data, kwLoc})
		}
		return n, nil
	}
	return nil, ErrNotSchema
}
