package registry

import (
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/jsonvalue"
)

func mustDoc(t *testing.T, s string) any {
	t.Helper()
	v, err := jsonvalue.DecodeLimited([]byte(s), jsonvalue.NumberLimits{})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func loadFixture(t *testing.T, name string) any {
	t.Helper()
	l, err := (&Loader{}).File(fixtures+name, "")
	if err != nil {
		t.Fatal(err)
	}
	return l.Doc
}

func newIdx(t *testing.T) *Index {
	t.Helper()
	idx, err := Build(true, "https://main.example/m.json")
	if err != nil {
		t.Fatal(err)
	}
	return idx
}

func TestRemappedFileReachableByBothAddresses(t *testing.T) {
	idx := newIdx(t)
	doc := mustDoc(t, `{"$id":"https://other.org/person.json","$defs":{"n":{"$id":"name.json"}}}`)
	if err := idx.AddExternal(doc, "https://example.com/a.json", "a.json"); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"https://example.com/a.json", "https://other.org/person.json", "https://other.org/name.json"} {
		if idx.Resources[k] == nil {
			t.Fatalf("%s not indexed", k)
		}
	}
	if idx.Resources["https://example.com/a.json"] != idx.Resources["https://other.org/person.json"] {
		t.Fatal("both addresses must reach the same resource")
	}
}

func TestFileWithoutAddressFoundByOwnID(t *testing.T) {
	idx := newIdx(t)
	if err := idx.AddExternal(mustDoc(t, `{"$id":"https://x.example/s.json"}`), "", "s.json"); err != nil {
		t.Fatal(err)
	}
	if idx.Resources["https://x.example/s.json"] == nil {
		t.Fatal("not indexed by $id")
	}
}

func TestDuplicateAddressNamesBothSources(t *testing.T) {
	idx := newIdx(t)
	if err := idx.AddExternal(loadFixture(t, "conflict/a.json"), "", "A-src"); err != nil {
		t.Fatal(err)
	}
	err := idx.AddExternal(loadFixture(t, "conflict/b.json"), "", "B-src")
	if err == nil {
		t.Fatal("different schemas at one address must fail")
	}
	for _, w := range []string{"https://example.com/conflict.json", "A-src", "B-src"} {
		if !strings.Contains(err.Error(), w) {
			t.Fatalf("error %q lacks %q", err, w)
		}
	}
}

func TestDeepEqualDuplicateAcceptedOnce(t *testing.T) {
	idx := newIdx(t)
	a := mustDoc(t, `{"$id":"https://x.example/d.json","minimum":1,"enum":[{"a":1,"b":2}]}`)
	b := mustDoc(t, `{"enum":[{"b":2,"a":1}],"minimum":1,"$id":"https://x.example/d.json"}`)
	if err := idx.AddExternal(a, "", "one"); err != nil {
		t.Fatal(err)
	}
	if err := idx.AddExternal(b, "", "two"); err != nil {
		t.Fatalf("deep-equal duplicate: %v", err)
	}
	if idx.Resources["https://x.example/d.json"].Source != "one" {
		t.Fatal("first source must stay")
	}
}
