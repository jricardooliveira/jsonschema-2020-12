package compile

import (
	"github.com/neoface/jsonschema/internal/registry"
)

// Doc is a schema document to be meta-checked and where it came from
// (empty for the main schema, else the supplied file label).
type Doc struct {
	Value  any
	Source string
}

// Check is one meta-schema evaluation to run: Meta applied to Doc.Value.
// compile cannot import eval, so the caller (the root package) runs it. No
// format assertion is involved: the evaluator only asserts format when asked.
type Check struct {
	Meta *Node
	Doc  Doc
}

// MetaChecks lists the evaluations that validate each of docs against its
// meta-schema. A custom meta-schema in ext is itself checked, along its chain
// (which ResolveMetaSchema bounds and requires to end in 2020-12). ext must
// hold the built-ins (registry.Index.AddBuiltins). Unsupported '$schema' and
// bad chains return ErrUnsupportedSchema / ErrMetaChain.
func MetaChecks(docs []Doc, ext *registry.Index) ([]Check, error) {
	var out []Check
	seen := map[string]bool{}
	for _, d := range docs {
		for cur := d; ; {
			meta, err := ResolveMetaSchema(ext, cur.Value)
			if err != nil {
				return nil, err
			}
			n, err := CompileWithRegistry(map[string]any{"$ref": meta}, StdinBaseURI, ext)
			if err != nil {
				return nil, err
			}
			out = append(out, Check{Meta: n, Doc: cur})
			if meta == Draft2020 || seen[meta] {
				break
			}
			seen[meta] = true
			r := ext.Resources[meta]
			cur = Doc{Value: r.Doc, Source: r.Source}
		}
	}
	return out, nil
}
