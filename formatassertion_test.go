package jsonschema_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neoface/jsonschema"
	"github.com/neoface/jsonschema/internal/keywords"
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

// A format check is charged against the per-call step budget (ADR-016 H): a
// long string under a small MaxSteps is a limit error in the mode, and the
// same budget is enough without the mode.
func TestFormatAssertionChargesSteps(t *testing.T) {
	doc := []byte(`"` + strings.Repeat("a", 1000*keywords.BytesPerStep) + `"`)
	for _, on := range []bool{false, true} {
		s, err := jsonschema.CompileWithOptions([]byte(`{"format":"hostname"}`), jsonschema.Options{FormatAssertion: on, MaxSteps: 100})
		if err != nil {
			t.Fatal(err)
		}
		_, err = s.Validate(doc)
		var le *jsonschema.LimitError
		if hit := errors.As(err, &le) && le.Limit == jsonschema.LimitSteps; hit != on {
			t.Errorf("mode %v: err %v", on, err)
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
