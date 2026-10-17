package compile

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/neoface/jsonschema/internal/keywords"
)

const (
	vCore = keywords.VocabCore
	vApp  = keywords.VocabApplicator
	vVal  = keywords.VocabValidation
	vFA   = keywords.VocabFormatAssertion
	meta  = "https://example.com/my-meta"
)

// core 8.1.2: a quoted entry is cut on a rune boundary, never mid-character.
func TestBoundRuneAware(t *testing.T) {
	got := bound("a" + strings.Repeat("é", maxEntryQuote))
	if !utf8.ValidString(got) || utf8.RuneCountInString(got) != maxEntryQuote+len("...") {
		t.Errorf("bound = %q", got)
	}
	if s := strings.Repeat("é", maxEntryQuote); bound(s) != s {
		t.Errorf("short entry changed: %q", bound(s))
	}
}

// 2020-12 core 8.1.2.
func TestParseVocabulary(t *testing.T) {
	type obj = map[string]any
	tests := []struct {
		name    string
		doc     any
		wantErr any // pointer to expected error type, nil for none
		active  []string
		inact   []string
		fa      bool
	}{
		{"missing $vocabulary: all active", obj{}, nil, keywords.StandardVocabularies(), nil, false},
		{"required known", obj{"$vocabulary": obj{vCore: true, vApp: true}}, nil, []string{vCore, vApp}, []string{vVal}, false},
		{"unlisted standard inactive", obj{"$vocabulary": obj{vCore: true, vApp: false}}, nil, []string{vCore, vApp}, []string{vVal}, false},
		{"optional unknown ignored", obj{"$vocabulary": obj{vCore: true, "https://example.com/v": false}}, nil, []string{vCore}, []string{"https://example.com/v"}, false},
		{"format-assertion required", obj{"$vocabulary": obj{vCore: true, vFA: true}}, nil, []string{vCore, vFA}, nil, true},
		{"format-assertion optional", obj{"$vocabulary": obj{vCore: true, vFA: false}}, nil, []string{vCore, vFA}, nil, false},
		{"required unknown", obj{"$vocabulary": obj{vCore: true, "https://example.com/v": true}}, new(*UnknownVocabularyError), nil, nil, false},
		{"core missing", obj{"$vocabulary": obj{vApp: true}}, new(*CoreVocabularyError), nil, nil, false},
		{"core optional", obj{"$vocabulary": obj{vCore: false}}, new(*CoreVocabularyError), nil, nil, false},
		{"not an object", obj{"$vocabulary": []any{}}, new(*MalformedVocabularyError), nil, nil, false},
		{"value not boolean", obj{"$vocabulary": obj{vCore: "yes"}}, new(*MalformedVocabularyError), nil, nil, false},
		{"key not absolute", obj{"$vocabulary": obj{vCore: true, "rel/uri": true}}, new(*MalformedVocabularyError), nil, nil, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			set, err := ParseVocabulary(meta, tc.doc)
			switch want := tc.wantErr.(type) {
			case nil:
				if err != nil {
					t.Fatal(err)
				}
			case **UnknownVocabularyError:
				if !errors.As(err, want) {
					t.Fatalf("got %v", err)
				}
				if err.Error() != "meta-schema "+meta+" requires unknown vocabulary https://example.com/v" {
					t.Errorf("message: %v", err)
				}
				return
			case **CoreVocabularyError:
				if !errors.As(err, want) {
					t.Fatalf("got %v", err)
				}
				return
			case **MalformedVocabularyError:
				if !errors.As(err, want) {
					t.Fatalf("got %v", err)
				}
				if tc.name == "value not boolean" && err.Error() != `meta-schema `+meta+`: invalid '$vocabulary' entry "`+vCore+`": value is not a boolean` {
					t.Errorf("message: %v", err)
				}
				return
			}
			for _, u := range tc.active {
				if !set.Active(u) {
					t.Errorf("%s should be active", u)
				}
			}
			for _, u := range tc.inact {
				if set.Active(u) {
					t.Errorf("%s should be inactive", u)
				}
			}
			if set.FormatAssertion != tc.fa {
				t.Errorf("FormatAssertion = %v", set.FormatAssertion)
			}
		})
	}
}
