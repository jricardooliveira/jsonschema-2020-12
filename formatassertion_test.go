package jsonschema_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neoface/jsonschema"
)

func faValid(t *testing.T, s *jsonschema.Schema, doc string) bool {
	t.Helper()
	r, err := s.Validate([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	return r.Valid()
}

func faMeta(t *testing.T, vocab, schema string) *jsonschema.Schema {
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

const (
	core = `"https://json-schema.org/draft/2020-12/vocab/core":true`
	fa   = `"https://json-schema.org/draft/2020-12/vocab/format-assertion"`
	// ponytail: format keyword is only active here through the assertion vocabulary.
	metaSchema = `{"$schema":"https://example.com/meta","format":"ipv4"}`
)

func TestFormatAssertionEnablement(t *testing.T) {
	const sch = `{"format":"date"}`
	off, _ := jsonschema.Compile([]byte(sch))
	on, err := jsonschema.CompileWithOptions([]byte(sch), jsonschema.Options{FormatAssertion: true})
	if err != nil {
		t.Fatal(err)
	}
	if !faValid(t, off, `"nope"`) || faValid(t, on, `"nope"`) || !faValid(t, on, `"2024-02-29"`) || !faValid(t, on, `7`) {
		t.Error("option on/off behaviour wrong")
	}
	if faValid(t, faMeta(t, `{`+core+`,`+fa+`:true}`, metaSchema), `"1.2.3.04"`) {
		t.Error("required format-assertion vocabulary must assert")
	}
	if !faValid(t, faMeta(t, `{`+core+`,`+fa+`:false}`, metaSchema), `"1.2.3.04"`) {
		t.Error("optional format-assertion vocabulary must not assert")
	}
}

func TestFormatAssertionDeferredAndUnknownPass(t *testing.T) {
	for _, f := range []string{"uri", "uri-reference", "iri", "iri-reference", "uri-template", "json-pointer", "relative-json-pointer", "regex", "no-such-format"} {
		s, err := jsonschema.CompileWithOptions([]byte(`{"format":"`+f+`"}`), jsonschema.Options{FormatAssertion: true})
		if err != nil {
			t.Fatal(err)
		}
		if !faValid(t, s, `"%%% not anything"`) {
			t.Errorf("format %s must pass", f)
		}
	}
}

// core 8.1.1: the meta-schema check never asserts format, in either mode.
func TestFormatAssertionMetaCheckAnnotationOnly(t *testing.T) {
	for _, on := range []bool{false, true} {
		if _, err := jsonschema.CompileWithOptions([]byte(`{"$id":"http://example.com/a b"}`), jsonschema.Options{FormatAssertion: on}); err != nil {
			t.Errorf("mode %v: %v", on, err)
		}
	}
}
