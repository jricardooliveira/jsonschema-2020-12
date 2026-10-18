package jsonschema_test

import (
	"os"
	"path/filepath"
	"testing"

	jsonschema "github.com/jricardooliveira/jsonschema-2020-12"
)

const vocabBase = "https://json-schema.org/draft/2020-12/vocab/"

func compileWithMeta(t *testing.T, vocab, schema string) *jsonschema.Schema {
	t.Helper()
	p := filepath.Join(t.TempDir(), "meta.json")
	meta := `{"$id":"https://example.com/meta","$schema":"https://json-schema.org/draft/2020-12/schema","$vocabulary":` + vocab + `}`
	if err := os.WriteFile(p, []byte(meta), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := jsonschema.CompileWithRegistry([]byte(schema), jsonschema.Options{},
		jsonschema.Registry{Files: []jsonschema.RegistryFile{{Path: p}}})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func valid(t *testing.T, s *jsonschema.Schema, doc string) bool {
	t.Helper()
	r, err := s.Validate([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	return r.Valid()
}

// core 8.1.2: keywords of an unlisted vocabulary are unknown; embedded
// resources use their own meta-schema's set.
func TestInactiveVocabularyAndEmbeddedResource(t *testing.T) {
	v := `{"` + vocabBase + `core":true,"` + vocabBase + `applicator":true}`
	s := compileWithMeta(t, v, `{"$schema":"https://example.com/meta","minimum":5}`)
	if !valid(t, s, `3`) {
		t.Error("minimum must be ignored without the validation vocabulary")
	}
	if s.FormatAssertionRequested() {
		t.Error("format assertion not requested")
	}
	s = compileWithMeta(t, v, `{"minimum":5,"properties":{"a":{"$id":"https://example.com/in","$schema":"https://example.com/meta","minimum":5}}}`)
	if valid(t, s, `3`) || !valid(t, s, `{"a":3}`) {
		t.Error("outer keeps validation, inner does not")
	}
}

// core 8.1.2: format-assertion required is recorded.
func TestFormatAssertionRecorded(t *testing.T) {
	v := `{"` + vocabBase + `core":true,"` + vocabBase + `format-assertion":true}`
	if !compileWithMeta(t, v, `{"$schema":"https://example.com/meta"}`).FormatAssertionRequested() {
		t.Error("format assertion should be recorded")
	}
}
