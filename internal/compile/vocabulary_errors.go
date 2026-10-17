package compile

import "fmt"

// UnknownVocabularyError: a meta-schema requires a vocabulary that is not recognised.
// Vocabulary is bounded like MalformedVocabularyError.Entry.
type UnknownVocabularyError struct{ Vocabulary, MetaSchema string }

func (e *UnknownVocabularyError) Error() string {
	return fmt.Sprintf("meta-schema %s requires unknown vocabulary %s", e.MetaSchema, e.Vocabulary)
}

// MalformedVocabularyError: '$vocabulary' of a meta-schema is not well formed.
// Entry is the (bounded) bad key, or "" when '$vocabulary' itself is bad.
type MalformedVocabularyError struct{ MetaSchema, Entry, Reason string }

func (e *MalformedVocabularyError) Error() string {
	if e.Entry == "" {
		return fmt.Sprintf("meta-schema %s: invalid '$vocabulary': %s", e.MetaSchema, e.Reason)
	}
	return fmt.Sprintf("meta-schema %s: invalid '$vocabulary' entry %q: %s", e.MetaSchema, e.Entry, e.Reason)
}

// UnsupportedSchemaError: '$schema' names another draft's meta-schema.
// Draft is the draft name taken from the URI (for example "draft-07").
// It matches ErrUnsupportedSchema under errors.Is.
type UnsupportedSchemaError struct{ Draft string }

func (e *UnsupportedSchemaError) Error() string {
	if e.Draft == "2020-12" { // the http:// spelling: only https identifies 2020-12
		return "unsupported $schema: only " + Draft2020 + " is supported"
	}
	return fmt.Sprintf("unsupported $schema: %s; only 2020-12 is supported", e.Draft)
}

// Is makes errors.Is(err, ErrUnsupportedSchema) true.
func (e *UnsupportedSchemaError) Is(target error) bool { return target == ErrUnsupportedSchema }

// CoreVocabularyError: the core vocabulary is missing or not required.
type CoreVocabularyError struct{ MetaSchema, Reason string }

func (e *CoreVocabularyError) Error() string {
	return fmt.Sprintf("meta-schema %s: core vocabulary %s", e.MetaSchema, e.Reason)
}
