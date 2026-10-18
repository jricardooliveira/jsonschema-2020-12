package compile

import (
	"errors"
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/registry"
)

func metaIdx(t *testing.T, supplied map[string]string) *registry.Index {
	t.Helper()
	idx, err := registry.Build(map[string]any{}, "urn:main")
	if err != nil {
		t.Fatal(err)
	}
	if err := idx.AddBuiltins(); err != nil {
		t.Fatal(err)
	}
	for addr, src := range supplied {
		doc, err := jsonvalue.Decode([]byte(src))
		if err != nil {
			t.Fatal(err)
		}
		if err := idx.AddExternal(doc, addr, addr+".json"); err != nil {
			t.Fatal(err)
		}
	}
	return idx
}

func schemaDoc(t *testing.T, src string) any {
	t.Helper()
	doc, err := jsonvalue.Decode([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

// Core 8.1.1: '$schema' selects the meta-schema; absent means 2020-12; an
// empty fragment is ignored; other drafts are unsupported.
func TestResolveMetaSchema(t *testing.T) {
	custom := map[string]string{
		"https://x.test/m1": `{"$id":"https://x.test/m1","$schema":"https://json-schema.org/draft/2020-12/schema"}`,
		"https://x.test/m2": `{"$id":"https://x.test/m2","$schema":"https://x.test/m1"}`,
		"https://x.test/m3": `{"$id":"https://x.test/m3"}`,
		"https://x.test/d7": `{"$id":"https://x.test/d7","$schema":"http://json-schema.org/draft-07/schema#"}`,
		"https://x.test/c1": `{"$id":"https://x.test/c1","$schema":"https://x.test/c2"}`,
		"https://x.test/c2": `{"$id":"https://x.test/c2","$schema":"https://x.test/c1"}`,
		"https://x.test/s":  `{"$id":"https://x.test/s","$schema":"https://x.test/s"}`,
	}
	idx := metaIdx(t, custom)
	cases := []struct {
		name, schema string
		want         string
		err          error
	}{
		{"absent", `{}`, Draft2020, nil},
		{"2020", `{"$schema":"https://json-schema.org/draft/2020-12/schema"}`, Draft2020, nil},
		{"2020 hash", `{"$schema":"https://json-schema.org/draft/2020-12/schema#"}`, Draft2020, nil},
		{"draft-07", `{"$schema":"http://json-schema.org/draft-07/schema#"}`, "", ErrUnsupportedSchema},
		{"custom", `{"$schema":"https://x.test/m1"}`, "https://x.test/m1", nil},
		{"custom chain", `{"$schema":"https://x.test/m2"}`, "https://x.test/m2", nil},
		{"custom no $schema", `{"$schema":"https://x.test/m3"}`, "https://x.test/m3", nil},
		{"chain to draft-07", `{"$schema":"https://x.test/d7"}`, "", ErrMetaChain},
		{"cycle", `{"$schema":"https://x.test/c1"}`, "", ErrMetaChain},
		{"self", `{"$schema":"https://x.test/s"}`, "", ErrMetaChain},
		{"boolean schema", `true`, Draft2020, nil},
	}
	for _, c := range cases {
		got, err := ResolveMetaSchema(idx, schemaDoc(t, c.schema))
		if !errors.Is(err, c.err) || got != c.want {
			t.Errorf("%s: got %q, %v; want %q, %v", c.name, got, err, c.want, c.err)
		}
		// The unsupported-draft message names the draft by design (story 190).
		if err != nil && !errors.Is(err, ErrUnsupportedSchema) && strings.Contains(err.Error(), "draft-07") {
			t.Errorf("%s: error echoes the declared URI: %v", c.name, err)
		}
	}
}
