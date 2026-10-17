package keywords

// The eight standard 2020-12 vocabulary URIs (core 8.1.2). The recognised set
// is exactly these: nothing else that registers a vocabulary widens it.
const (
	VocabCore             = "https://json-schema.org/draft/2020-12/vocab/core"
	VocabApplicator       = "https://json-schema.org/draft/2020-12/vocab/applicator"
	VocabUnevaluated      = "https://json-schema.org/draft/2020-12/vocab/unevaluated"
	VocabValidation       = "https://json-schema.org/draft/2020-12/vocab/validation"
	VocabMetaData         = "https://json-schema.org/draft/2020-12/vocab/meta-data"
	VocabFormatAnnotation = "https://json-schema.org/draft/2020-12/vocab/format-annotation"
	VocabFormatAssertion  = "https://json-schema.org/draft/2020-12/vocab/format-assertion"
	VocabContent          = "https://json-schema.org/draft/2020-12/vocab/content"
)

var standardVocabularies = []string{
	VocabCore, VocabApplicator, VocabUnevaluated, VocabValidation,
	VocabMetaData, VocabFormatAnnotation, VocabFormatAssertion, VocabContent,
}

// StandardVocabularies returns a copy of the recognised vocabulary URIs.
func StandardVocabularies() []string {
	return append([]string(nil), standardVocabularies...)
}

// IsStandardVocabulary reports whether uri is a recognised 2020-12 vocabulary.
func IsStandardVocabulary(uri string) bool {
	for _, v := range standardVocabularies {
		if v == uri {
			return true
		}
	}
	return false
}
