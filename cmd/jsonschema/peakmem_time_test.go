//go:build darwin

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/perfdata"
)

// TestCLIPeakMemoryTimeL builds the CLI into a temporary directory and runs
// `/usr/bin/time -l <cli> validate -o basic schema 5.5MB.json` five times,
// logging the "maximum resident set size" line (bytes) of each run, for
// docs/benchmarks.md. Information only, never asserted; runs only with
// -peakmem, and only on macOS (BSD time's -l).
//
//	go test -count=1 -run 'TestCLIPeakMemoryTimeL$' -v ./cmd/jsonschema -args -peakmem
func TestCLIPeakMemoryTimeL(t *testing.T) {
	if !*peakMem {
		t.Skip("pass -args -peakmem to measure")
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "jsonschema")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	schema, err := perfdata.Fixture("order_schema.json")
	if err != nil {
		t.Fatal(err)
	}
	schemaPath := filepath.Join(dir, "schema.json")
	docPath := filepath.Join(dir, "5.5MB.json")
	doc := perfdata.ValidOrder(perfdata.LargeItems)
	if err := os.WriteFile(schemaPath, schema, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(docPath, doc, 0o600); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 5; i++ {
		cmd := exec.Command("/usr/bin/time", "-l", bin, "validate", "-o", "basic", schemaPath, docPath)
		var stderr strings.Builder
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("run %d: %v\n%s", i, err, stderr.String())
		}
		for _, line := range strings.Split(stderr.String(), "\n") {
			if strings.Contains(line, "maximum resident set size") {
				t.Logf("run %d (%d bytes): %s", i, len(doc), strings.TrimSpace(line))
			}
		}
	}
}
