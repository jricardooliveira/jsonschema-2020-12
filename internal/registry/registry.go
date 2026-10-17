package registry

import "sort"

// BuiltinSource is the source label of the embedded meta-schemas in errors.
const BuiltinSource = "built-in meta-schema"

// AddBuiltins indexes the embedded meta-schemas under their $id. Call it
// before adding supplied documents: AddExternal then accepts a supplied file at
// a built-in address only when it is deep-equal to the embedded one (the
// embedded copy stays indexed) and refuses a different one with an error that
// names the address and the file, quoting no content.
func (idx *Index) AddBuiltins() error {
	docs, err := Builtins()
	if err != nil {
		return err
	}
	ids := make([]string, 0, len(docs))
	for id := range docs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if err := idx.AddExternal(docs[id], id, BuiltinSource); err != nil {
			return err
		}
	}
	return nil
}
