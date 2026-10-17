//go:build unix

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/perfdata"
)

// TestCLIPeakMemory reports the peak resident memory of one CLI run per
// document, for docs/benchmarks.md. Information only, never asserted: it runs
// only with -peakmem. Each run is a fresh child process (this test binary
// re-executed) that calls run once and reports getrusage's maxrss (bytes on
// macOS, KiB on Linux). The "floor" row is the child with a tiny document.
//
//	go test -count=1 -run TestCLIPeakMemory -v ./cmd/jsonschema -args -peakmem
func TestCLIPeakMemory(t *testing.T) {
	if !*peakMem {
		t.Skip("pass -args -peakmem to measure")
	}
	schema, err := perfdata.Fixture("order_schema.json")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "schema.json")
	if err := os.WriteFile(schemaPath, schema, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name string
		doc  []byte
		want int
	}{
		{"floor", []byte(`{"id": 1, "items": []}`), exitValid},
		{"0.5MB", perfdata.ValidOrder(perfdata.MediumItems), exitValid},
		{"5.5MB", perfdata.ValidOrder(perfdata.LargeItems), exitValid},
		{"Errors", perfdata.ErrorOrder(), exitInvalid},
	} {
		docPath := filepath.Join(dir, c.name+".json")
		if err := os.WriteFile(docPath, c.doc, 0o600); err != nil {
			t.Fatal(err)
		}
		spec := schemaPath + "|" + docPath + "|" + strconv.Itoa(c.want)
		out, err := exec.Command(os.Args[0], "-test.run=^TestCLIPeakMemoryChild$", "-peakmem-child="+spec).CombinedOutput()
		if err != nil {
			t.Fatalf("%s: %v\n%s", c.name, err, out)
		}
		for _, line := range strings.Split(string(out), "\n") {
			if rss, ok := strings.CutPrefix(line, "maxrss "); ok {
				t.Logf("%s (%d bytes): maxrss %s", c.name, len(c.doc), rss)
			}
		}
	}
}

// TestCLIPeakMemoryChild is the child half of TestCLIPeakMemory.
func TestCLIPeakMemoryChild(t *testing.T) {
	if *peakMemChild == "" {
		t.Skip("child of TestCLIPeakMemory only")
	}
	p := strings.Split(*peakMemChild, "|")
	want, _ := strconv.Atoi(p[2])
	if code := run([]string{"validate", p[0], p[1]}, nil, io.Discard, io.Discard); code != want {
		t.Fatalf("exit = %d, want %d", code, want)
	}
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("maxrss %d\n", ru.Maxrss)
}
