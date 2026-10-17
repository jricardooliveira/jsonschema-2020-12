// Package all pulls in every 2020-12 vocabulary package so their keywords
// register themselves. Import it once, for side effects.
package all

import (
	_ "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords/applicator"
	_ "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords/content"
	_ "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords/core"
	_ "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords/formatannotation"
	_ "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords/metadata"
	_ "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords/unevaluated"
	_ "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords/validation"
)
