package metadata

import "github.com/neoface/jsonschema/internal/keywords"

// VocabularyURI is the 2020-12 meta-data vocabulary.
const VocabularyURI = "https://json-schema.org/draft/2020-12/vocab/meta-data"

// Each keyword only annotates its own schema value (validation 9.1).
func init() {
	for _, n := range []string{"title", "description", "default", "deprecated", "readOnly", "writeOnly", "examples"} {
		keywords.RegisterAnnotationOnly(n, VocabularyURI)
	}
}
