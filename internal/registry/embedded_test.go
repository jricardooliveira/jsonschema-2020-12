package registry

import (
	"strings"
	"testing"
)

const draft = "https://json-schema.org/draft/2020-12/"

var builtinIDs = []string{
	draft + "schema",
	draft + "meta/core", draft + "meta/applicator", draft + "meta/unevaluated",
	draft + "meta/validation", draft + "meta/meta-data", draft + "meta/format-annotation",
	draft + "meta/format-assertion", draft + "meta/content",
}

// Core 8.1.1 / 8.2.1: every built-in $id resolves from the index, offline.
func TestEveryBuiltinResolvesOffline(t *testing.T) {
	idx := newIdx(t)
	if err := idx.AddBuiltins(); err != nil {
		t.Fatal(err)
	}
	for _, id := range builtinIDs {
		r, ok := idx.Resources[id]
		if !ok {
			t.Errorf("%s not indexed", id)
			continue
		}
		if r.Source != BuiltinSource {
			t.Errorf("%s: source %q", id, r.Source)
		}
	}
	if _, ok := idx.Resources[draft+"meta/core"].Anchors["meta"]; !ok {
		t.Error("meta/core must expose the dynamic anchor meta")
	}
}

// As published, only the 2020-12 meta-schema declares '$vocabulary'; the
// eight vocabulary meta-schemas do not.
func TestOnlyRootMetaSchemaDeclaresVocabulary(t *testing.T) {
	docs, err := Builtins()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range builtinIDs {
		_, has := docs[id].(map[string]any)["$vocabulary"]
		if has != (id == draft+"schema") {
			t.Errorf("%s: has $vocabulary = %v", id, has)
		}
	}
}

// Core 8.2.1: an equal supplied file at a built-in address is accepted.
func TestEqualSuppliedBuiltinAccepted(t *testing.T) {
	idx := newIdx(t)
	if err := idx.AddBuiltins(); err != nil {
		t.Fatal(err)
	}
	docs, _ := Builtins()
	if err := idx.AddExternal(docs[draft+"schema"], draft+"schema", "mine.json"); err != nil {
		t.Fatal(err)
	}
	if idx.Resources[draft+"schema"].Source != BuiltinSource {
		t.Error("the embedded copy must stay indexed")
	}
}

// A different supplied file at a built-in address is refused, naming address and file.
func TestDifferentSuppliedBuiltinRefused(t *testing.T) {
	idx := newIdx(t)
	if err := idx.AddBuiltins(); err != nil {
		t.Fatal(err)
	}
	err := idx.AddExternal(mustDoc(t, `{"$id":"`+draft+`schema","type":"string"}`), "", "evil.json")
	if err == nil {
		t.Fatal("want error")
	}
	if !strings.Contains(err.Error(), draft+"schema") || !strings.Contains(err.Error(), "evil.json") {
		t.Errorf("error must name address and file: %v", err)
	}
	if strings.Contains(err.Error(), "string") {
		t.Errorf("error must not quote content: %v", err)
	}
}
