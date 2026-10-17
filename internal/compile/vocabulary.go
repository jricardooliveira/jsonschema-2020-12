package compile

import (
	"net/url"
	"sort"

	"github.com/neoface/jsonschema/internal/keywords"
)

// VocabularySet is the set of vocabularies active for a schema resource.
// It is never modified after ParseVocabulary returns.
type VocabularySet struct {
	active map[string]bool
	// FormatAssertion is true when format-assertion was listed as required.
	FormatAssertion bool
}

// Active reports whether the vocabulary uri is active.
func (s *VocabularySet) Active(uri string) bool { return s.active[uri] }

const maxEntryQuote = 80

func bound(s string) string {
	if len(s) > maxEntryQuote {
		return s[:maxEntryQuote] + "..."
	}
	return s
}

// ParseVocabulary reads '$vocabulary' from the meta-schema document doc
// (named metaSchema in errors) and returns its active set (core 8.1.2).
// Without '$vocabulary' every standard vocabulary is active. It fails closed:
// any problem is an error, never a fallback to "all active".
func ParseVocabulary(metaSchema string, doc any) (*VocabularySet, error) {
	m, _ := doc.(map[string]any)
	raw, present := m["$vocabulary"]
	set := &VocabularySet{active: map[string]bool{}}
	if !present {
		for _, u := range keywords.StandardVocabularies() {
			set.active[u] = true
		}
		return set, nil
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, &MalformedVocabularyError{MetaSchema: metaSchema, Reason: "value is not an object"}
	}
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if u, err := url.Parse(k); err != nil || !u.IsAbs() {
			return nil, &MalformedVocabularyError{metaSchema, bound(k), "key is not an absolute URI"}
		}
		if _, ok := obj[k].(bool); !ok {
			return nil, &MalformedVocabularyError{metaSchema, bound(k), "value is not a boolean"}
		}
	}
	if req, ok := obj[keywords.VocabCore]; !ok {
		return nil, &CoreVocabularyError{metaSchema, "is missing"}
	} else if !req.(bool) {
		return nil, &CoreVocabularyError{metaSchema, "must be required (true)"}
	}
	for _, k := range keys {
		if keywords.IsStandardVocabulary(k) {
			set.active[k] = true
			if k == keywords.VocabFormatAssertion && obj[k].(bool) {
				set.FormatAssertion = true
			}
		} else if obj[k].(bool) {
			return nil, &UnknownVocabularyError{Vocabulary: bound(k), MetaSchema: metaSchema}
		}
	}
	return set, nil
}
