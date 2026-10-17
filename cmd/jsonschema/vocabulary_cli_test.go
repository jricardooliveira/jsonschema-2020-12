package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const vocabCLI = "../../testdata/vocabulary/"

type vocabCLICase struct {
	Name          string          `json:"name"`
	Meta          string          `json:"meta"`
	Schema        json.RawMessage `json:"schema"`
	Want          string          `json:"want"`
	ErrorContains string          `json:"errorContains"`
	Docs          []struct {
		Doc   json.RawMessage `json:"doc"`
		Valid bool            `json:"valid"`
	} `json:"docs"`
}

// Core 8.1.1 and 8.1.2: the same cases through the CLI. A refused schema is
// exit 2 with a value-free stderr line; documents give exit 0 or 1.
func TestCLIVocabularyCases(t *testing.T) {
	files, _ := filepath.Glob(vocabCLI + "*.json")
	if len(files) != 3 {
		t.Fatalf("want 3 case files, found %v", files)
	}
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var cases []vocabCLICase
		if err := json.Unmarshal(raw, &cases); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		for _, c := range cases {
			t.Run(filepath.Base(f)+"/"+c.Name, func(t *testing.T) {
				dir := t.TempDir()
				schema := filepath.Join(dir, "schema.json")
				if err := os.WriteFile(schema, c.Schema, 0o600); err != nil {
					t.Fatal(err)
				}
				args := []string{"validate"}
				if c.Meta != "" {
					args = append(args, "--ref", vocabCLI+"metaschemas/"+c.Meta)
				}
				args = append(args, schema)
				if c.Want == "error" {
					doc := filepath.Join(dir, "doc.json")
					if err := os.WriteFile(doc, []byte(`1`), 0o600); err != nil {
						t.Fatal(err)
					}
					code, out, errOut := do(append(args, doc), "")
					wantErr(t, c.Name, code, out, errOut)
					if !strings.Contains(errOut, c.ErrorContains) {
						t.Errorf("stderr %q lacks %q", errOut, c.ErrorContains)
					}
					return
				}
				for _, d := range c.Docs {
					doc := filepath.Join(dir, "doc.json")
					if err := os.WriteFile(doc, d.Doc, 0o600); err != nil {
						t.Fatal(err)
					}
					want := exitInvalid
					if d.Valid {
						want = exitValid
					}
					if code, out, errOut := do(append(args, doc), ""); code != want || errOut != "" {
						t.Errorf("doc %s: code=%d out=%q err=%q, want %d", d.Doc, code, out, errOut, want)
					}
				}
			})
		}
	}
}
