package jsonschema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type specGroup struct {
	Spec        string          `json:"spec"`
	Description string          `json:"description"`
	Schema      json.RawMessage `json:"schema"`
	Tests       []struct {
		Description string          `json:"description"`
		Data        json.RawMessage `json:"data"`
		Valid       bool            `json:"valid"`
	} `json:"tests"`
}

// TestSpecCases runs every testdata/**/*.json file through Compile/Validate.
func TestSpecCases(t *testing.T) {
	files, err := filepath.Glob("testdata/*/*.json")
	if err != nil || len(files) == 0 {
		t.Fatalf("no testdata files: %v", err)
	}
	for _, f := range files {
		if filepath.Base(filepath.Dir(f)) == "metaschema" {
			continue // own case shape and runner: metaschema_test.go
		}
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var groups []specGroup
		if err := json.Unmarshal(raw, &groups); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		for _, g := range groups {
			if g.Spec == "" {
				t.Errorf("%s: %q cites no spec section", f, g.Description)
			}
			// A group rejected by the meta-schema fails here: invalid-schema
			// cases belong in testdata/metaschema (metaschema_test.go).
			s, err := Compile(g.Schema)
			if err != nil {
				t.Errorf("%s: %s: compile: %v", f, g.Description, err)
				continue
			}
			for _, c := range g.Tests {
				t.Run(filepath.Base(f)+"/"+g.Description+"/"+c.Description, func(t *testing.T) {
					r, err := s.Validate(c.Data)
					if err != nil || r.Valid() != c.Valid {
						t.Errorf("got valid=%v err=%v, want %v (%s)", r.Valid(), err, c.Valid, g.Spec)
					}
				})
			}
		}
	}
}
