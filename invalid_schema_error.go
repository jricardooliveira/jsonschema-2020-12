package jsonschema

import (
	"fmt"
	"strings"
)

// InvalidSchemaError is returned by Compile when a schema (the main one or a
// supplied one it reaches) fails its meta-schema. It is never a validation
// result. It holds no schema or instance values. An unsupported '$schema' or
// a bad meta-schema chain is a different, plain error from Compile, also
// never a result.
type InvalidSchemaError struct {
	// Source is the supplied file label, or "" for the main schema.
	Source string
	// SchemaLocation is the JSON Pointer of the failing place in the schema.
	SchemaLocation string
	// Keyword is the failing meta-schema keyword.
	Keyword string
}

func (e *InvalidSchemaError) Error() string {
	src := "schema"
	if e.Source != "" {
		src = "schema " + e.Source
	}
	return fmt.Sprintf("%s: invalid against its meta-schema at %q: rule %q failed", src, e.SchemaLocation, e.Keyword)
}

// newInvalidSchemaError picks the deepest failing unit (value-free by design).
func newInvalidSchemaError(source, instLoc, kwLoc string) *InvalidSchemaError {
	kw := kwLoc
	if i := strings.LastIndexByte(kwLoc, '/'); i >= 0 {
		kw = kwLoc[i+1:]
	}
	return &InvalidSchemaError{Source: source, SchemaLocation: instLoc, Keyword: kw}
}
