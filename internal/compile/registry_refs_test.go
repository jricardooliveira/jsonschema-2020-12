package compile

import (
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords/core"
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

// supplied is a registry holding docs, each (address, source, json).
func supplied(t *testing.T, docs ...[3]string) *registry.Index {
	t.Helper()
	idx, err := registry.Build(true, "https://unused.example/m.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range docs {
		if err := idx.AddExternal(decode(t, d[2]), d[0], d[1]); err != nil {
			t.Fatal(err)
		}
	}
	return idx
}

func compileWith(t *testing.T, main, base string, ext *registry.Index) (*Node, error) {
	t.Helper()
	return CompileWithRegistry(decode(t, main), base, ext)
}

// refTarget returns the node the $ref at the root of n points at.
func refTarget(t *testing.T, n *Node) *Node {
	t.Helper()
	for _, it := range n.Items {
		if r, ok := it.Data.(*core.Ref); ok {
			return r.Target.(*Node)
		}
	}
	t.Fatal("no $ref")
	return nil
}

func TestRefByOwnIDAndMappedURI(t *testing.T) {
	ext := supplied(t, [3]string{"https://example.com/a.json", "a.json", `{"$id":"https://other.org/person.json","type":"object"}`})
	for _, ref := range []string{"https://example.com/a.json", "https://other.org/person.json"} {
		n, err := compileWith(t, `{"$ref":"`+ref+`"}`, "file:///work/main.json", ext)
		if err != nil {
			t.Fatalf("%s: %v", ref, err)
		}
		if refTarget(t, n) == nil {
			t.Fatal("target not set")
		}
	}
}

func TestRelativeRefInsideRemappedFileUsesItsOwnID(t *testing.T) {
	ext := supplied(t,
		[3]string{"https://example.com/a.json", "a.json", `{"$id":"https://other.org/dir/person.json","properties":{"n":{"$ref":"name.json"}}}`},
		[3]string{"", "name.json", `{"$id":"https://other.org/dir/name.json","type":"string"}`})
	if _, err := compileWith(t, `{"$ref":"https://example.com/a.json"}`, "file:///work/main.json", ext); err != nil {
		t.Fatal(err)
	}
	// Resolved against the mapped address it would be https://example.com/name.json: not supplied.
	bad := supplied(t, [3]string{"https://example.com/a.json", "a.json", `{"properties":{"n":{"$ref":"name.json"}}}`},
		[3]string{"", "name.json", `{"$id":"https://other.org/dir/name.json"}`})
	if _, err := compileWith(t, `{"$ref":"https://example.com/a.json"}`, "file:///work/main.json", bad); err == nil {
		t.Fatal("relative ref must resolve against the file's base")
	}
}

func TestAnchorAndPointerInSuppliedFile(t *testing.T) {
	ext := supplied(t, [3]string{"https://x.example/s.json", "s.json", `{"$defs":{"a":{"$anchor":"A","type":"string"}}}`})
	for _, ref := range []string{"#A", "#/$defs/a"} {
		if _, err := compileWith(t, `{"$ref":"https://x.example/s.json`+ref+`"}`, "file:///m.json", ext); err != nil {
			t.Fatalf("%s: %v", ref, err)
		}
	}
}

func TestDefaultFileBaseResolvesNeighbourOnlyIfListed(t *testing.T) {
	main := `{"$ref":"address.json"}`
	if _, err := compileWith(t, main, "file:///work/main.json", nil); err == nil || !strings.Contains(err.Error(), "unresolved reference") {
		t.Fatalf("unlisted file:// neighbour must be unresolved, got %v", err)
	}
	ext := supplied(t, [3]string{"file:///work/address.json", "address.json", `{"type":"string"}`})
	if _, err := compileWith(t, main, "file:///work/main.json", ext); err != nil {
		t.Fatal(err)
	}
}

func TestUnresolvedAbsoluteRef(t *testing.T) {
	ext := supplied(t, [3]string{"https://x.example/a.json", "a.json", `true`})
	_, err := compileWith(t, `{"$ref":"https://x.example/missing.json"}`, "file:///m.json", ext)
	if err == nil || !strings.Contains(err.Error(), "unresolved reference") {
		t.Fatalf("got %v", err)
	}
}

// A broken reference inside a reached supplied file fails at compile time;
// a supplied file nobody reaches is not checked.
func TestReachedResourceResolvedAtCompileTime(t *testing.T) {
	broken := `{"properties":{"p":{"$ref":"nowhere.json"}}}`
	ext := supplied(t, [3]string{"https://x.example/b.json", "b.json", broken})
	if _, err := compileWith(t, `{"$ref":"https://x.example/b.json"}`, "file:///m.json", ext); err == nil {
		t.Fatal("reached file with unresolved ref must fail at compile")
	}
	if _, err := compileWith(t, `true`, "file:///m.json", ext); err != nil {
		t.Fatalf("unreached file must not be checked: %v", err)
	}
}

func TestCycleAcrossSuppliedFilesTerminates(t *testing.T) {
	ext := supplied(t,
		[3]string{"https://x.example/a.json", "a.json", `{"properties":{"b":{"$ref":"b.json"}}}`},
		[3]string{"https://x.example/b.json", "b.json", `{"properties":{"a":{"$ref":"a.json"}}}`})
	if _, err := compileWith(t, `{"$ref":"https://x.example/a.json"}`, "file:///m.json", ext); err != nil {
		t.Fatal(err)
	}
}
