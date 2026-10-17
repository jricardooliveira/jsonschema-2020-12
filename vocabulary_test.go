package jsonschema_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	jsonschema "github.com/neoface/jsonschema"
)

type vocabCase struct {
	Name            string          `json:"name"`
	Spec            string          `json:"spec"`
	Meta            string          `json:"meta"` // file in testdata/vocabulary/metaschemas, supplied through the registry
	Schema          json.RawMessage `json:"schema"`
	Want            string          `json:"want"` // ok | error
	ErrorContains   string          `json:"errorContains"`
	FormatAssertion bool            `json:"formatAssertion"`
	Docs            []vocabDoc      `json:"docs"`
}

type vocabDoc struct {
	Doc   json.RawMessage `json:"doc"`
	Valid bool            `json:"valid"`
}

// Core 8.1.1 and 8.1.2: vocabulary handling through the library, with custom
// meta-schemas supplied through the registry.
func TestVocabularyCases(t *testing.T) {
	files, _ := filepath.Glob("testdata/vocabulary/*.json")
	if len(files) != 3 {
		t.Fatalf("want 3 case files, found %v", files)
	}
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var cases []vocabCase
		if err := json.Unmarshal(raw, &cases); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		for _, c := range cases {
			if c.Spec == "" || !strings.HasPrefix(c.Spec, "core 8.") {
				t.Errorf("%s: case %q must cite a core section", f, c.Name)
			}
			t.Run(filepath.Base(f)+"/"+c.Name, func(t *testing.T) { runVocabCase(t, c) })
		}
	}
}

func runVocabCase(t *testing.T, c vocabCase) {
	var reg jsonschema.Registry
	if c.Meta != "" {
		reg.Files = []jsonschema.RegistryFile{{Path: filepath.Join("testdata/vocabulary/metaschemas", c.Meta)}}
	}
	s, err := jsonschema.CompileWithRegistry(c.Schema, jsonschema.Options{}, reg)
	if c.Want == "error" {
		if err == nil {
			t.Fatal("want a compile error")
		}
		if !strings.Contains(err.Error(), c.ErrorContains) {
			t.Fatalf("error %q lacks %q", err, c.ErrorContains)
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if s.FormatAssertionRequested() != c.FormatAssertion {
		t.Errorf("FormatAssertionRequested = %v", s.FormatAssertionRequested())
	}
	for _, d := range c.Docs {
		r, err := s.Validate(d.Doc)
		if err != nil {
			t.Fatal(err)
		}
		if r.Valid() != d.Valid {
			t.Errorf("doc %s: valid = %v, want %v", d.Doc, r.Valid(), d.Valid)
		}
	}
}

// Core 8.1.2: the error types are distinct and found with errors.As.
func TestVocabularyErrorTypes(t *testing.T) {
	compile := func(meta, id string) error {
		_, err := jsonschema.CompileWithRegistry([]byte(`{"$schema":"https://example.com/meta/`+id+`"}`), jsonschema.Options{},
			jsonschema.Registry{Files: []jsonschema.RegistryFile{{Path: "testdata/vocabulary/metaschemas/" + meta}}})
		return err
	}
	var unk *jsonschema.UnknownVocabularyError
	if err := compile("unknown_required.json", "unknown-required"); !errors.As(err, &unk) ||
		unk.Vocabulary != "https://example.com/vocab/money" || unk.MetaSchema != "https://example.com/meta/unknown-required" {
		t.Errorf("unknown vocabulary: %v", err)
	}
	var core *jsonschema.CoreVocabularyError
	if err := compile("no_core.json", "no-core"); !errors.As(err, &core) {
		t.Errorf("core: %v", err)
	}
	var uns *jsonschema.UnsupportedSchemaError
	_, err := jsonschema.Compile([]byte(`{"$schema":"http://json-schema.org/draft-07/schema#"}`))
	if !errors.As(err, &uns) || uns.Draft != "draft-07" {
		t.Errorf("unsupported: %v", err)
	}
}

// Core 8.1.1: '$schema' produces no annotation in a passing basic report.
func TestSchemaKeywordSilent(t *testing.T) {
	s, err := jsonschema.Compile([]byte(`{"$schema":"https://json-schema.org/draft/2020-12/schema","type":"string"}`))
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Validate([]byte(`"x"`))
	if err != nil {
		t.Fatal(err)
	}
	out, err := r.Output("basic")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "$schema") {
		t.Errorf("report has a $schema annotation: %s", out)
	}
}
