package jsonschema_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	jsonschema "github.com/jricardooliveira/jsonschema-2020-12"
)

// compileRefTo compiles a main schema that only $refs a registry resource
// whose own '$schema' is given, with the vocabulary meta-schemas supplied.
func compileRefTo(t *testing.T, schemaURI string) error {
	t.Helper()
	p := filepath.Join(t.TempDir(), "ext.json")
	ext := `{"$id":"https://example.com/ext","$schema":"` + schemaURI + `","type":"string"}`
	if err := os.WriteFile(p, []byte(ext), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := jsonschema.CompileWithRegistry([]byte(`{"$ref":"https://example.com/ext"}`), jsonschema.Options{},
		jsonschema.Registry{Files: []jsonschema.RegistryFile{
			{Path: p},
			{Path: "testdata/vocabulary/metaschemas/unknown_required.json"},
		}})
	return err
}

// core 8.1.1, 8.1.2: a registry resource reached through $ref is refused with
// its own error type, not a flattened message.
func TestRefReachedResourceVocabularyErrors(t *testing.T) {
	var unk *jsonschema.UnknownVocabularyError
	err := compileRefTo(t, "https://example.com/meta/unknown-required")
	if !errors.As(err, &unk) || unk.Vocabulary != "https://example.com/vocab/money" {
		t.Errorf("unknown vocabulary: %v", err)
	}
	if err != nil && !strings.Contains(err.Error(), "meta-schema https://example.com/meta/unknown-required requires unknown vocabulary https://example.com/vocab/money") {
		t.Errorf("message: %v", err)
	}
	var uns *jsonschema.UnsupportedSchemaError
	if err := compileRefTo(t, "http://json-schema.org/draft-07/schema#"); !errors.As(err, &uns) || uns.Draft != "draft-07" {
		t.Errorf("unsupported: %v", err)
	}
}
