package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/perfdata"
)

// benchCLI times one in-process run of the CLI on a document written to a
// temporary file once, before the timer starts. Output is discarded.
func benchCLI(b *testing.B, doc []byte, wantExit int, extra ...string) {
	schema, err := perfdata.Fixture("order_schema.json")
	if err != nil {
		b.Fatal(err)
	}
	dir := b.TempDir()
	schemaPath := filepath.Join(dir, "schema.json")
	docPath := filepath.Join(dir, "doc.json")
	if err := os.WriteFile(schemaPath, schema, 0o600); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(docPath, doc, 0o600); err != nil {
		b.Fatal(err)
	}
	args := append([]string{"validate"}, extra...)
	args = append(args, schemaPath, docPath)
	b.SetBytes(int64(len(doc)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if code := run(args, nil, io.Discard, io.Discard); code != wantExit {
			b.Fatalf("exit = %d, want %d", code, wantExit)
		}
	}
}

func BenchmarkCLI(b *testing.B) {
	b.Run("0.5MB", func(b *testing.B) {
		benchCLI(b, perfdata.ValidOrder(perfdata.MediumItems), exitValid)
	})
	b.Run("5.5MB", func(b *testing.B) {
		benchCLI(b, perfdata.ValidOrder(perfdata.LargeItems), exitValid)
	})
	b.Run("Errors", func(b *testing.B) {
		benchCLI(b, perfdata.ErrorOrder(), exitInvalid)
	})
	b.Run("Verbose0.5MB", func(b *testing.B) {
		benchCLI(b, perfdata.ValidOrder(perfdata.MediumItems), exitValid, "-output", "verbose")
	})
}
