package jsonschema

import "github.com/jricardooliveira/jsonschema-2020-12/internal/compile"

// Vocabulary and '$schema' errors returned by Compile. Detect them with
// errors.As. They carry only URIs and the bad entry's key, never instance values.
type (
	// UnknownVocabularyError: a meta-schema requires an unrecognised vocabulary.
	UnknownVocabularyError = compile.UnknownVocabularyError
	// MalformedVocabularyError: a meta-schema's '$vocabulary' is not well formed.
	MalformedVocabularyError = compile.MalformedVocabularyError
	// CoreVocabularyError: the core vocabulary is missing or not required.
	CoreVocabularyError = compile.CoreVocabularyError
	// UnsupportedSchemaError: '$schema' names another draft's meta-schema.
	UnsupportedSchemaError = compile.UnsupportedSchemaError
)
