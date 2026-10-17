package compile

import (
	"github.com/jricardooliveira/jsonschema-2020-12/internal/registry"
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

// MetaChecks validates each of docs against its meta-schema by calling run
// once per evaluation, stopping at the first error run returns. A custom
// meta-schema in ext is itself checked along its chain (which
// ResolveMetaSchema bounds and requires to end in 2020-12), from the 2020-12
// end down, so each custom meta-schema is checked before it is compiled and
// used. ext must hold the built-ins (registry.Index.AddBuiltins). Unsupported
// '$schema' and bad chains return ErrUnsupportedSchema / ErrMetaChain.
func MetaChecks(docs []Doc, ext *registry.Index, run func(Check) error) error {
	type link struct {
		doc  Doc
		meta string
	}
	seen := map[string]bool{}
	for _, d := range docs {
		var chain []link
		for cur := d; ; {
			meta, err := ResolveMetaSchema(ext, cur.Value)
			if err != nil {
				return err
			}
			chain = append(chain, link{cur, meta})
			if meta == Draft2020 || seen[meta] {
				break
			}
			seen[meta] = true
			r := ext.Resources[meta]
			cur = Doc{Value: r.Doc, Source: r.Source}
		}
		for i := len(chain) - 1; i >= 0; i-- {
			// Format assertion is forced off: meta-checks never assert format.
			var n *Node
			if chain[i].meta == Draft2020 {
				// The 2020-12 meta-schema refers only to embedded resources, which
				// a caller cannot override: reuse the shared compiled one.
				set, err := sharedBuiltin()
				if err != nil {
					return err
				}
				n = set.meta
			} else {
				var err error
				_, n, err = compileDocumentOpts(map[string]any{"$ref": chain[i].meta}, StdinBaseURI, ext, 0, false, true)
				if err != nil {
					return err
				}
			}
			if err := run(Check{Meta: n, Doc: chain[i].doc}); err != nil {
				return err
			}
		}
	}
	return nil
}
