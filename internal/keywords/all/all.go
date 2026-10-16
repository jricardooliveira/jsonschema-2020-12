// Package all pulls in every 2020-12 vocabulary package so their keywords
// register themselves. Import it once, for side effects.
package all

import (
	_ "github.com/neoface/jsonschema/internal/keywords/applicator"
	_ "github.com/neoface/jsonschema/internal/keywords/content"
	_ "github.com/neoface/jsonschema/internal/keywords/core"
	_ "github.com/neoface/jsonschema/internal/keywords/formatannotation"
	_ "github.com/neoface/jsonschema/internal/keywords/metadata"
	_ "github.com/neoface/jsonschema/internal/keywords/unevaluated"
	_ "github.com/neoface/jsonschema/internal/keywords/validation"
)
