package jsonschema_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	jsonschema "github.com/jricardooliveira/jsonschema-2020-12"
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

// Core 8.1.2 (AC4): under a meta-schema listing only core and applicator,
// 'minimum' is an unknown keyword: 3 passes and '/minimum' is an annotation
// with value 5 in the passing basic report.
func TestUnlistedVocabularyKeywordAnnotated(t *testing.T) {
	s, err := jsonschema.CompileWithRegistry([]byte(`{"$schema":"https://example.com/meta/core-applicator","minimum":5}`), jsonschema.Options{},
		jsonschema.Registry{Files: []jsonschema.RegistryFile{{Path: "testdata/vocabulary/metaschemas/core_applicator_only.json"}}})
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Validate([]byte(`3`))
	if err != nil {
		t.Fatal(err)
	}
	if !r.Valid() {
		t.Fatal("minimum must not reject 3 without the validation vocabulary")
	}
	out, err := r.Output("basic")
	if err != nil {
		t.Fatal(err)
	}
	var rep struct {
		Valid       bool `json:"valid"`
		Annotations []struct {
			KeywordLocation string          `json:"keywordLocation"`
			Annotation      json.RawMessage `json:"annotation"`
		} `json:"annotations"`
	}
	if err := json.Unmarshal(out, &rep); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	found := false
	for _, a := range rep.Annotations {
		if a.KeywordLocation == "/minimum" && string(a.Annotation) == "5" {
			found = true
		}
	}
	if !rep.Valid || !found {
		t.Errorf("want a passing report with a '/minimum' annotation of 5: %s", out)
	}
}

// Core 8.1.1: only the https URI identifies 2020-12; the http variant is
// refused with a message that is not self-contradictory.
func TestHTTP2020SchemaRefused(t *testing.T) {
	_, err := jsonschema.Compile([]byte(`{"$schema":"http://json-schema.org/draft/2020-12/schema"}`))
	want := "unsupported $schema: only https://json-schema.org/draft/2020-12/schema is supported"
	var uns *jsonschema.UnsupportedSchemaError
	if !errors.As(err, &uns) || !strings.Contains(err.Error(), want) {
		t.Errorf("got %v, want %q", err, want)
	}
}

// Core 8.1.2: the vocabulary URI quoted in UnknownVocabularyError is bounded.
func TestUnknownVocabularyErrorBounded(t *testing.T) {
	long := "https://example.com/" + strings.Repeat("v", 5000)
	p := filepath.Join(t.TempDir(), "meta.json")
	meta := `{"$id":"https://example.com/meta","$schema":"https://json-schema.org/draft/2020-12/schema","$vocabulary":{"` +
		vocabBase + `core":true,"` + long + `":true}}`
	if err := os.WriteFile(p, []byte(meta), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := jsonschema.CompileWithRegistry([]byte(`{"$schema":"https://example.com/meta"}`), jsonschema.Options{},
		jsonschema.Registry{Files: []jsonschema.RegistryFile{{Path: p}}})
	var unk *jsonschema.UnknownVocabularyError
	if !errors.As(err, &unk) {
		t.Fatalf("got %v", err)
	}
	if len(err.Error()) > 300 || strings.Contains(err.Error(), long) {
		t.Errorf("unbounded vocabulary URI in message (%d bytes)", len(err.Error()))
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

// core 8.1.2: vocabulary sets are fixed at Compile and only read during
// evaluation, so one schema with a restricted-vocabulary meta-schema and an
// embedded 2020-12 resource gives identical results from many goroutines
// (run with -race).
func TestVocabularyConcurrentValidate(t *testing.T) {
	v := `{"` + vocabBase + `core":true,"` + vocabBase + `applicator":true}`
	s := compileWithMeta(t, v, `{"$schema":"https://example.com/meta","minimum":5,"properties":{
		"a":{"$id":"https://example.com/in","$schema":"https://json-schema.org/draft/2020-12/schema","minimum":5}}}`)
	docs := map[string]bool{`3`: true, `{"a":3}`: false, `{"a":7}`: true}
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 50 {
				for d, want := range docs {
					r, err := s.Validate([]byte(d))
					if err != nil {
						t.Error(err)
						return
					}
					if r.Valid() != want {
						t.Errorf("doc %s: valid = %v, want %v", d, r.Valid(), want)
					}
				}
			}
		}()
	}
	wg.Wait()
}
