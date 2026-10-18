package compile

import (
	"strings"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

// vocabAt returns the vocabulary set in effect at schema location loc, or nil
// when every vocabulary is active (core 8.1.2). The set belongs to the schema
// resource holding loc: its own '$schema' decides, else the enclosing
// resource's set is inherited, else everything is active. Computed at compile
// time only; sets are never changed afterwards.
func (r *resolver) vocabAt(loc string) (*VocabularySet, error) {
	return r.vocabFor(r.baseAt(loc))
}

func (r *resolver) vocabFor(base string) (*VocabularySet, error) {
	if set, ok := r.vocabs[base]; ok {
		return set, nil
	}
	var set *VocabularySet
	if res := r.resources[base]; res != nil {
		v, _ := navigate(r.doc, splitPtr(res.Pointer))
		name, err := declared(v)
		if err != nil {
			return nil, err
		}
		switch {
		case name != "" && name != Draft2020:
			meta, err := ResolveMetaSchema(r.sh.ext, v)
			if err != nil {
				return nil, err
			}
			if set, err = ParseVocabulary(meta, r.sh.ext.Resources[meta].Doc); err != nil {
				return nil, err
			}
		case name == "" && res.Pointer != "":
			// A resource without '$schema' inherits its parent's set.
			parent := res.Pointer[:strings.LastIndexByte(res.Pointer, '/')]
			if set, err = r.vocabAt(parent); err != nil {
				return nil, err
			}
		}
	}
	if set != nil && set.FormatAssertion {
		r.sh.formatAssertion = true
	}
	r.vocabs[base] = set
	return set, nil
}

// allows reports whether keyword kw may be evaluated under set. A keyword of
// a standard vocabulary the set does not list is treated as unknown.
func (set *VocabularySet) allows(kw keywords.Keyword) bool {
	if set == nil || !keywords.IsStandardVocabulary(kw.Vocabulary) {
		return true
	}
	return set.Active(kw.Vocabulary)
}
