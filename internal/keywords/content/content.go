package content

import "github.com/neoface/jsonschema/internal/keywords"

// VocabularyURI is the 2020-12 content vocabulary.
const VocabularyURI = "https://json-schema.org/draft/2020-12/vocab/content"

// Annotation only (validation 8). contentSchema is kept as a raw value and is
// never compiled or applied to the instance.
func init() {
	for _, n := range []string{"contentEncoding", "contentMediaType", "contentSchema"} {
		keywords.RegisterAnnotationOnly(n, VocabularyURI)
	}
}
