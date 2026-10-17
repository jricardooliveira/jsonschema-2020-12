// Package speccases runs the spec-derived JSON cases in testdata/<area>/<keyword>.json
// through the public library.
package speccases

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12"
)

// Group is one schema with its instances and the spec section it derives from.
type Group struct {
	Spec        string          `json:"spec"`
	Description string          `json:"description"`
	Schema      json.RawMessage `json:"schema"`
	Tests       []struct {
		Description string          `json:"description"`
		Data        json.RawMessage `json:"data"`
		Valid       bool            `json:"valid"`
	} `json:"tests"`
}

// Load reads one case file. A group without a spec citation is an error.
func Load(path string) ([]Group, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var groups []Group
	if err := json.Unmarshal(raw, &groups); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	for _, g := range groups {
		if strings.TrimSpace(g.Spec) == "" {
			return nil, fmt.Errorf("%s: %q cites no spec section", path, g.Description)
		}
	}
	return groups, nil
}

// leaks reports whether msg contains the instance's compact serialisation.
// Instances of fewer than 4 bytes are skipped: they would match by chance.
func leaks(msg string, data []byte) bool {
	var b bytes.Buffer
	if json.Compact(&b, data) != nil || b.Len() < 4 {
		return false
	}
	return strings.Contains(msg, b.String())
}

// Run executes every case file matching glob as subtests of t.
func Run(t *testing.T, glob string) {
	t.Helper()
	files, err := filepath.Glob(glob)
	if err != nil || len(files) == 0 {
		t.Fatalf("no case files match %q: %v", glob, err)
	}
	for _, f := range files {
		groups, err := Load(f)
		if err != nil {
			t.Errorf("%v", err)
			continue
		}
		for _, g := range groups {
			s, err := jsonschema.Compile(g.Schema)
			if err != nil {
				t.Errorf("%s: %s: compile: %v", f, g.Description, err)
				continue
			}
			for _, c := range g.Tests {
				t.Run(filepath.Base(f)+"/"+g.Description+"/"+c.Description, func(t *testing.T) {
					r, err := s.Validate(c.Data)
					if err != nil {
						if leaks(err.Error(), c.Data) {
							t.Errorf("error message contains the instance: %v", err)
						}
						t.Errorf("error: %v (%s)", err, g.Spec)
						return
					}
					if r.Valid() != c.Valid {
						t.Errorf("got valid=%v, want %v (%s)", r.Valid(), c.Valid, g.Spec)
					}
				})
			}
		}
	}
}
