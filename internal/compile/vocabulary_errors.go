package compile

import "fmt"

// UnknownVocabularyError: a meta-schema requires a vocabulary that is not recognised.
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

// CoreVocabularyError: the core vocabulary is missing or not required.
type CoreVocabularyError struct{ MetaSchema, Reason string }

func (e *CoreVocabularyError) Error() string {
	return fmt.Sprintf("meta-schema %s: core vocabulary %s", e.MetaSchema, e.Reason)
}
