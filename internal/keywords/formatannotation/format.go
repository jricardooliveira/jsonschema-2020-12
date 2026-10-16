package formatannotation

import "github.com/neoface/jsonschema/internal/keywords"

// VocabularyURI is the 2020-12 format-annotation vocabulary.
const VocabularyURI = "https://json-schema.org/draft/2020-12/vocab/format-annotation"

// format annotates its value and never fails (validation 7.2.1); unknown
// formats are never an error.
func init() { keywords.RegisterAnnotationOnly("format", VocabularyURI) }
