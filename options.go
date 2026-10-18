package jsonschema

import "github.com/jricardooliveira/jsonschema-2020-12/internal/compile"

// baseURI picks the root base URI: BaseURI, else SchemaPath's file URI, else
// the fixed stdin placeholder. An invalid BaseURI is an error, never ignored.
func (o Options) baseURI() (string, error) {
	return compile.RootBaseURI(o.BaseURI, o.SchemaPath)
}
