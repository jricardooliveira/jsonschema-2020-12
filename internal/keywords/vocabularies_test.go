package keywords

import "testing"

// 2020-12 core 8.1.2: the eight standard vocabularies are recognised.
func TestStandardVocabulariesRecognised(t *testing.T) {
	got := StandardVocabularies()
	if len(got) != 8 {
		t.Fatalf("want 8 vocabularies, got %d", len(got))
	}
	for _, u := range got {
		if !IsStandardVocabulary(u) {
			t.Errorf("%s not recognised", u)
		}
	}
	if IsStandardVocabulary("https://example.com/vocab/money") {
		t.Error("unknown vocabulary recognised")
	}
}
