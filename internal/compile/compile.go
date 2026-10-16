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
		ctx := &keywords.CompileContext{}
		for _, name := range names {
			kw, ok := keywords.Lookup(name)
			if !ok {
				continue // ponytail: unknown keywords ignored; annotations come later
			}
			data, err := kw.Compile(ctx, s[name])
			if err != nil {
				return nil, err
			}
			n.Items = append(n.Items, Item{kw, data, loc + "/" + escape.Replace(name)})
		}
		return n, nil
	}
	return nil, ErrNotSchema
}
