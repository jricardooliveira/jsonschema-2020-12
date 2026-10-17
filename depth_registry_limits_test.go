package jsonschema

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func nestArr(n int) string { return strings.Repeat("[", n) + strings.Repeat("]", n) }

func TestDepthLimitSchemaAndDocument(t *testing.T) {
	const max = 5
	o := Options{MaxDepth: max}
	// schema: depth max is fine, max+1 is a LimitError
	if _, err := CompileWithOptions([]byte(`{"enum":[`+nestArr(max-2)+`]}`), o); err != nil {
		t.Fatalf("schema at cap: %v", err)
	}
	var le *LimitError
	_, err := CompileWithOptions([]byte(`{"enum":[`+nestArr(max-1)+`]}`), o)
	if !errors.As(err, &le) || le.Limit != LimitDepth || le.Value != max {
		t.Fatalf("schema over cap: %v", err)
	}
	s, err := CompileWithOptions([]byte(`true`), o)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Validate([]byte(nestArr(max))); err != nil {
		t.Fatalf("doc at cap: %v", err)
	}
	if _, err := s.Validate([]byte(nestArr(max - 1))); err != nil {
		t.Fatalf("doc below cap: %v", err)
	}
	le = nil
	if _, err := s.Validate([]byte(nestArr(max + 1))); !errors.As(err, &le) || le.Limit != LimitDepth || le.Value != max {
		t.Fatalf("doc over cap: %v", err)
	}
}

// encoding/json stops at DefaultMaxDepth, so a higher setting never turns
// into a "syntax error" or a valid result: it is still a LimitError.
func TestDepthAboveStdlibCeilingIsLimitError(t *testing.T) {
	s, err := CompileWithOptions([]byte(`true`), Options{MaxDepth: DefaultMaxDepth * 2})
	if err != nil {
		t.Fatal(err)
	}
	var le *LimitError
	if _, err := s.Validate([]byte(nestArr(DefaultMaxDepth + 1))); !errors.As(err, &le) || le.Limit != LimitDepth {
		t.Fatalf("got %v", err)
	}
}

func TestRegistryTotalAndFileLimits(t *testing.T) {
	dir := t.TempDir()
	body := `{"$id":"https://x.test/%s","type":"string"}`
	var files []RegistryFile
	size := 0
	for _, n := range []string{"a", "b"} {
		b := strings.Replace(body, "%s", n, 1)
		size = len(b)
		p := filepath.Join(dir, n+".json")
		if err := os.WriteFile(p, []byte(b), 0o600); err != nil {
			t.Fatal(err)
		}
		files = append(files, RegistryFile{Path: p})
	}
	compile := func(r Registry) error {
		_, err := CompileWithRegistry([]byte(`true`), Options{}, r)
		return err
	}
	if err := compile(Registry{Files: files, MaxTotalBytes: 2 * size}); err != nil {
		t.Fatalf("total at cap: %v", err)
	}
	var le *LimitError
	err := compile(Registry{Files: files, MaxTotalBytes: 2*size - 1})
	if !errors.As(err, &le) || le.Limit != LimitRegistrySize || le.Value != 2*size-1 {
		t.Fatalf("total over cap: %v", err)
	}
	// per file: explicit cap, and the default follows MaxInputBytes
	if err := compile(Registry{Files: files, MaxFileBytes: size}); err != nil {
		t.Fatalf("file at cap: %v", err)
	}
	le = nil
	err = compile(Registry{Files: files, MaxFileBytes: size - 1})
	if !errors.As(err, &le) || le.Limit != LimitInputSize {
		t.Fatalf("file over cap: %v", err)
	}
	le = nil
	_, err = CompileWithRegistry([]byte(`true`), Options{MaxInputBytes: size - 1 + 8}, Registry{Files: files})
	if err != nil {
		t.Fatalf("schema smaller than cap, files within: %v", err)
	}
	_, err = CompileWithRegistry([]byte(`true`), Options{MaxInputBytes: size - 1}, Registry{Files: files})
	if !errors.As(err, &le) || le.Limit != LimitInputSize {
		t.Fatalf("file over MaxInputBytes: %v", err)
	}
	if err := compile(Registry{Files: files, MaxTotalBytes: -1}); err == nil {
		t.Fatal("negative total accepted")
	}
}
