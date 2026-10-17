package jsonschema

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type metaCase struct {
	Name             string          `json:"name"`
	Spec             string          `json:"spec"`
	Schema           json.RawMessage `json:"schema"`
	Registry         []metaRegFile   `json:"registry"`
	Want             string          `json:"want"` // ok | invalid_schema | error
	Location         *string         `json:"location"`
	Source           string          `json:"source"` // expected file label suffix
	ErrorContains    string          `json:"errorContains"`
	ErrorNotContains string          `json:"errorNotContains"`
}

type metaRegFile struct {
	URI    string          `json:"uri"`
	Doc    json.RawMessage `json:"doc"`
	CopyOf string          `json:"copyOf"` // repo file whose content is supplied
}

// Core 8.1.1 / 8.2.1: meta-schema checks at load time, through the public API.
// Registry files are written to a temp dir, named after the last URI segment.
func TestMetaSchemaCases(t *testing.T) {
	files, _ := filepath.Glob("testdata/metaschema/*.json")
	if len(files) != 4 {
		t.Fatalf("want 4 case files, found %v", files)
	}
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var cases []metaCase
		if err := json.Unmarshal(raw, &cases); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		for _, c := range cases {
			t.Run(filepath.Base(f)+"/"+c.Name, func(t *testing.T) { runMetaCase(t, c) })
		}
	}
}

func runMetaCase(t *testing.T, c metaCase) {
	t.Logf("derives from %s", c.Spec)
	dir := t.TempDir()
	var reg Registry
	for _, r := range c.Registry {
		data := []byte(r.Doc)
		if r.CopyOf != "" {
			var err error
			if data, err = os.ReadFile(r.CopyOf); err != nil {
				t.Fatal(err)
			}
		}
		p := filepath.Join(dir, r.URI[strings.LastIndex(r.URI, "/")+1:]+".json")
		if err := os.WriteFile(p, data, 0o600); err != nil {
			t.Fatal(err)
		}
		reg.Files = append(reg.Files, RegistryFile{URI: r.URI, Path: p})
	}
	_, err := CompileWithRegistry([]byte(c.Schema), Options{}, reg)
	var ise *InvalidSchemaError
	isInvalid := errors.As(err, &ise)
	switch c.Want {
	case "ok":
		if err != nil {
			t.Fatalf("want ok, got %v", err)
		}
		return
	case "invalid_schema":
		if !isInvalid {
			t.Fatalf("want *InvalidSchemaError, got %v", err)
		}
		if c.Location != nil && ise.SchemaLocation != *c.Location {
			t.Errorf("location %q, want %q", ise.SchemaLocation, *c.Location)
		}
		if c.Source != "" && !strings.Contains(ise.Source, c.Source) {
			t.Errorf("source %q, want it to contain %q", ise.Source, c.Source)
		}
	case "error":
		if err == nil {
			t.Fatal("want an error, got none")
		}
		if isInvalid {
			t.Errorf("want a plain error, got InvalidSchemaError: %v", err)
		}
	default:
		t.Fatalf("bad want %q", c.Want)
	}
	if c.ErrorContains != "" && !strings.Contains(err.Error(), c.ErrorContains) {
		t.Errorf("error %q lacks %q", err, c.ErrorContains)
	}
	if c.ErrorNotContains != "" && strings.Contains(err.Error(), c.ErrorNotContains) {
		t.Errorf("error %q must not contain %q", err, c.ErrorNotContains)
	}
}
