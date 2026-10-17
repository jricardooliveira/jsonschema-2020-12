package compile_test

import (
	"errors"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/eval"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	_ "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords/all"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/registry"
)

func decode(t *testing.T, s string) any {
	t.Helper()
	v, err := jsonvalue.DecodeLimited([]byte(s), jsonvalue.NumberLimits{})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func idx(t *testing.T) *registry.Index {
	t.Helper()
	i, err := registry.Build(true, "https://registry.invalid/unused")
	if err != nil {
		t.Fatal(err)
	}
	if err := i.AddBuiltins(); err != nil {
		t.Fatal(err)
	}
	return i
}

var errFailed = errors.New("meta-check failed")

// allPass runs every check and reports whether all passed.
func allPass(t *testing.T, ext *registry.Index, docs []compile.Doc) bool {
	t.Helper()
	_, ok := firstFailure(t, ext, docs)
	return ok
}

// firstFailure returns the source of the first failing document, or ok.
func firstFailure(t *testing.T, ext *registry.Index, docs []compile.Doc) (string, bool) {
	t.Helper()
	var failed string
	err := compile.MetaChecks(docs, ext, func(c compile.Check) error {
		ev, err := eval.EvaluateFull(c.Meta, c.Doc.Value, 1_000_000, eval.DefaultMaxAnnotations)
		if err != nil {
			return err
		}
		if !ev.Valid {
			failed = c.Doc.Source
			return errFailed
		}
		return nil
	})
	if err != nil && err != errFailed {
		t.Fatal(err)
	}
	return failed, err == nil
}

// Core 8.1.1: an invalid custom meta-schema is reported before it is used.
func TestCustomMetaSchemaCheckedFirst(t *testing.T) {
	ext := idx(t)
	if err := ext.AddExternal(decode(t, `{"$id":"https://x.test/m","type":"nope"}`), "", "m.json"); err != nil {
		t.Fatal(err)
	}
	src, ok := firstFailure(t, ext, []compile.Doc{{Value: decode(t, `{"$schema":"https://x.test/m"}`)}})
	if ok || src != "m.json" {
		t.Errorf("got ok=%v source %q, want the meta-schema m.json to fail", ok, src)
	}
}

// Core 8.2.4.2 / 8.1.1: "$dynamicRef": "#meta" reaches nested subschemas.
func TestMetaCheckNestedThroughDynamicRef(t *testing.T) {
	for src, want := range map[string]bool{
		`{"type":"string"}`: true,
		`{"properties":{"a":{"items":{"allOf":[{"type":"nope"}]}}}}`:                  false,
		`{"$defs":{"x":{"minLength":-1}}}`:                                            false,
		`{"$schema":"https://json-schema.org/draft/2020-12/schema#","type":"object"}`: true,
	} {
		doc := decode(t, src)
		if got := allPass(t, idx(t), []compile.Doc{{Value: doc}}); got != want {
			t.Errorf("%s: got %v want %v", src, got, want)
		}
	}
}

// Story 189: only reached supplied schemas are checked; unreferenced ones are ignored.
func TestReachedDocsOnly(t *testing.T) {
	ext := idx(t)
	if err := ext.AddExternal(decode(t, `{"$id":"https://x.test/ok","title":5}`), "", "ok.json"); err != nil {
		t.Fatal(err)
	}
	if err := ext.AddExternal(decode(t, `{"$id":"https://x.test/unused","title":5}`), "", "unused.json"); err != nil {
		t.Fatal(err)
	}
	_, docs, err := compile.CompileReached(decode(t, `{"$ref":"https://x.test/ok"}`), compile.StdinBaseURI, ext)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 2 || docs[1].Source != "ok.json" {
		t.Fatalf("docs: %+v", docs)
	}
	if allPass(t, ext, docs) {
		t.Error("reached invalid supplied schema must fail")
	}
	if !allPass(t, ext, docs[:1]) {
		t.Error("main alone is valid")
	}
}
