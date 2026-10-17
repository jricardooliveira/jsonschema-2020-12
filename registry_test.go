package jsonschema

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
)

const regDir = "testdata/registry/"

func compileFixture(t *testing.T, main string, reg Registry) (*Schema, error) {
	t.Helper()
	data, err := os.ReadFile(regDir + main)
	if err != nil {
		t.Fatal(err)
	}
	return CompileWithRegistry(data, Options{SchemaPath: regDir + main}, reg)
}

func mustValidate(t *testing.T, s *Schema, doc string) bool {
	t.Helper()
	r, err := s.Validate([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	return r.Valid()
}

// core 8.2: a mapped file is reached by its mapped URI.
func TestRegistryFileMapped(t *testing.T) {
	s, err := compileFixture(t, "file/main.json", Registry{Files: []RegistryFile{{URI: "https://example.com/person.json", Path: regDir + "file/person.json"}}})
	if err != nil {
		t.Fatal(err)
	}
	if !mustValidate(t, s, `{"name":"a"}`) || mustValidate(t, s, `{}`) {
		t.Fatal("wrong verdicts")
	}
}

// A mapped file with a different $id is reached by both addresses.
func TestRegistryFileBothAddresses(t *testing.T) {
	reg := Registry{Files: []RegistryFile{{URI: "https://example.com/s.json", Path: regDir + "file/self_id.json"}}}
	for _, ref := range []string{"https://example.com/s.json", "https://other.org/self.json"} {
		s, err := CompileWithRegistry([]byte(`{"$ref":"`+ref+`"}`), Options{}, reg)
		if err != nil {
			t.Fatal(ref, err)
		}
		if !mustValidate(t, s, `1`) || mustValidate(t, s, `"x"`) {
			t.Fatal("wrong verdicts for", ref)
		}
	}
}

func TestRegistryFileWithoutURIFoundByID(t *testing.T) {
	s, err := CompileWithRegistry([]byte(`{"$ref":"https://other.org/self.json"}`), Options{}, Registry{Files: []RegistryFile{{Path: regDir + "file/self_id.json"}}})
	if err != nil {
		t.Fatal(err)
	}
	if !mustValidate(t, s, `3`) {
		t.Fatal("expected valid")
	}
}

func TestRegistryDirMapped(t *testing.T) {
	s, err := compileFixture(t, "dir/main.json", Registry{Dirs: []RegistryDir{{BaseURI: "https://example.com/schemas/", Path: regDir + "dir/schemas"}}})
	if err != nil {
		t.Fatal(err)
	}
	if !mustValidate(t, s, `{"address":{"street":"x"}}`) || mustValidate(t, s, `{"address":{}}`) {
		t.Fatal("wrong verdicts")
	}
}

func TestRegistryUnresolvedIsError(t *testing.T) {
	_, err := compileFixture(t, "unresolved/main.json", Registry{Files: []RegistryFile{{Path: regDir + "file/self_id.json"}}})
	if err == nil || !strings.Contains(err.Error(), "unresolved reference") {
		t.Fatalf("got %v", err)
	}
}

func TestRegistryUnlistedFileURLUnresolved(t *testing.T) {
	_, err := CompileWithRegistry([]byte(`{"$ref":"file:///etc/passwd"}`), Options{}, Registry{})
	if err == nil {
		t.Fatal("must be unresolved")
	}
}

func TestRegistryLoadErrors(t *testing.T) {
	cases := map[string]Registry{
		"missing":    {Files: []RegistryFile{{Path: regDir + "nope.json"}}},
		"broken":     {Files: []RegistryFile{{Path: regDir + "broken/bad.json"}}},
		"relative":   {Files: []RegistryFile{{URI: "rel.json", Path: regDir + "file/person.json"}}},
		"negative":   {MaxFileBytes: -1},
		"dir-nobase": {Dirs: []RegistryDir{{Path: regDir + "dir/schemas"}}},
	}
	for name, reg := range cases {
		if _, err := CompileWithRegistry([]byte(`true`), Options{}, reg); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

func TestRegistryDuplicateAddress(t *testing.T) {
	_, err := CompileWithRegistry([]byte(`true`), Options{}, Registry{Files: []RegistryFile{
		{Path: regDir + "conflict/a.json"}, {Path: regDir + "conflict/b.json"}}})
	if err == nil || !strings.Contains(err.Error(), "https://example.com/conflict.json") ||
		!strings.Contains(err.Error(), "a.json") || !strings.Contains(err.Error(), "b.json") {
		t.Fatalf("got %v", err)
	}
	same := RegistryFile{Path: regDir + "conflict/a.json"}
	if _, err := CompileWithRegistry([]byte(`true`), Options{}, Registry{Files: []RegistryFile{same, same}}); err != nil {
		t.Fatalf("deep-equal duplicate: %v", err)
	}
}

func TestRegistryLimitsRaisable(t *testing.T) {
	f := RegistryFile{Path: regDir + "file/person.json"}
	var le *jsonvalue.LimitError
	_, err := CompileWithRegistry([]byte(`true`), Options{}, Registry{Files: []RegistryFile{f}, MaxFileBytes: 5})
	if !errors.As(err, &le) || le.Limit != "input size" { // story 193: the limit error names the limit, not the file
		t.Fatalf("per-file: %v", err)
	}
	_, err = CompileWithRegistry([]byte(`true`), Options{}, Registry{Files: []RegistryFile{f, {Path: regDir + "file/self_id.json"}}, MaxTotalBytes: 100})
	if !errors.As(err, &le) {
		t.Fatalf("total: %v", err)
	}
	if _, err = CompileWithRegistry([]byte(`true`), Options{}, Registry{Files: []RegistryFile{f}, MaxFileBytes: 1 << 20, MaxTotalBytes: 1 << 20}); err != nil {
		t.Fatal(err)
	}
}

// The registry is fixed at compile time: rewriting the supplied files on disk
// and changing the caller's Files and Dirs afterwards leave the verdicts of a
// Schema shared across goroutines unchanged.
func TestRegistrySchemaConcurrent(t *testing.T) {
	tmp := t.TempDir()
	person := filepath.Join(tmp, "person.json")
	addrDir := filepath.Join(tmp, "schemas")
	address := filepath.Join(addrDir, "common", "address.json")
	for _, c := range [][2]string{{regDir + "file/person.json", person}, {regDir + "dir/schemas/common/address.json", address}} {
		data, err := os.ReadFile(c[0])
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(c[1]), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(c[1], data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	files := []RegistryFile{{URI: "https://example.com/person.json", Path: person}}
	dirs := []RegistryDir{{BaseURI: "https://example.com/schemas/", Path: addrDir}}
	main := `{"properties":{"p":{"$ref":"https://example.com/person.json"},"a":{"$ref":"https://example.com/schemas/common/address.json"}}}`
	s, err := CompileWithRegistry([]byte(main), Options{}, Registry{Files: files, Dirs: dirs})
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{person, address} {
		if err := os.WriteFile(p, []byte(`true`), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	files[0] = RegistryFile{URI: "https://example.com/other.json", Path: address}
	dirs[0] = RegistryDir{BaseURI: "https://changed.example/", Path: tmp}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !mustValidate(t, s, `{"p":{"name":"a"},"a":{"street":"x"}}`) ||
				mustValidate(t, s, `{"p":{}}`) || mustValidate(t, s, `{"a":{}}`) {
				t.Error("verdicts changed after the registry sources changed")
			}
		}()
	}
	wg.Wait()
}

const relRefSchema = `{"$defs":{"e":{"$id":"http://x/y/e.json"}},"$ref":"e.json"}`

// core 9.1: the base URI decides how a relative $ref resolves.
func TestBaseURIOption(t *testing.T) {
	if _, err := CompileWithOptions([]byte(relRefSchema), Options{BaseURI: "http://x/y/s.json"}); err != nil {
		t.Errorf("override: %v", err)
	}
	if _, err := CompileWithOptions([]byte(relRefSchema), Options{BaseURI: "http://x/z/s.json"}); err == nil {
		t.Error("expected unresolved reference")
	}
	if _, err := Compile([]byte(relRefSchema)); err == nil {
		t.Error("default (stdin placeholder) base must not resolve e.json")
	}
}

func TestBaseURIOptionRejectsBadValues(t *testing.T) {
	for _, bad := range []string{"rel/path", "http://x/a#frag", "http://x/%zz"} {
		if _, err := CompileWithOptions([]byte(`true`), Options{BaseURI: bad}); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestSchemaPathGivesFileBase(t *testing.T) {
	s := `{"$id":"file:///d/e.json","$ref":"file:///d/e.json#/nope"}`
	if _, err := CompileWithOptions([]byte(s), Options{SchemaPath: "/d/s.json"}); err == nil {
		t.Error("expected error")
	}
	rel := `{"$defs":{"e":{"$id":"file:///d/e.json"}},"$ref":"e.json"}`
	if _, err := CompileWithOptions([]byte(rel), Options{SchemaPath: "/d/s.json"}); err != nil {
		t.Error(err)
	}
}
