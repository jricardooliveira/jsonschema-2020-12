package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"testing"
)

// goldenCapture turns the capture on. It is a one-time step on the unchanged
// v1.1 code: the files are never regenerated or edited afterwards, and the
// capture refuses to overwrite them.
var goldenCapture = flag.Bool("golden-capture", false, "capture testdata/golden/*.json (once, from v1.1; refuses to overwrite)")

const (
	goldenSmallFile = "testdata/golden/cli_v1_1.json"
	goldenLargeFile = "testdata/golden/cli_v1_1_large.json"
)

// TestGoldenCapture runs every case twice and writes the golden files. Any
// case whose two runs differ stops the capture and is listed.
func TestGoldenCapture(t *testing.T) {
	if !*goldenCapture {
		t.Skip("opt-in: go test ./cmd/jsonschema -run TestGoldenCapture -golden-capture")
	}
	paths := []string{filepath.Join(goldenRoot, goldenSmallFile), filepath.Join(goldenRoot, goldenLargeFile)}
	for _, p := range paths {
		if _, err := os.Stat(p); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("%s exists (or cannot be checked): the capture never overwrites golden files", p)
		}
	}
	if err := os.MkdirAll(filepath.Dir(paths[0]), 0o755); err != nil {
		t.Fatal(err)
	}
	cases, err := goldenCases()
	if err != nil {
		t.Fatal(err)
	}
	bin, err := goldenBuild(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	type job struct {
		c     goldenCase
		shape string
		i     int
	}
	var jobs []job
	for _, c := range cases {
		for _, s := range goldenShapes {
			jobs = append(jobs, job{c, s, len(jobs)})
		}
	}
	recs := make([]goldenRecord, len(jobs))
	var mu sync.Mutex
	var unstable, failed []string
	ch := make(chan job)
	var wg sync.WaitGroup
	for w := 0; w < runtime.NumCPU(); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range ch {
				a, err1 := goldenRun(bin, j.c, j.shape)
				b, err2 := goldenRun(bin, j.c, j.shape)
				mu.Lock()
				switch {
				case err1 != nil || err2 != nil:
					failed = append(failed, a.ID)
				case !reflect.DeepEqual(a, b):
					unstable = append(unstable, a.ID)
				default:
					recs[j.i] = a
				}
				mu.Unlock()
			}
		}()
	}
	for _, j := range jobs {
		ch <- j
	}
	close(ch)
	wg.Wait()
	if len(failed) > 0 {
		t.Fatalf("could not run %d cases: %v", len(failed), failed)
	}
	if len(unstable) > 0 {
		t.Fatalf("STOP: %d cases give different output on two runs: %v", len(unstable), unstable)
	}
	var small, large goldenFile
	for _, r := range recs {
		if r.isLarge() {
			large.Cases = append(large.Cases, r)
		} else {
			small.Cases = append(small.Cases, r)
		}
	}
	for i, gf := range []*goldenFile{&small, &large} {
		gf.Header.Provenance = goldenProvenance
		gf.Header.GoVersion = runtime.Version()
		gf.Header.Cases = len(gf.Cases)
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", " ")
		if err := enc.Encode(gf); err != nil {
			t.Fatal(err)
		}
		f, err := os.OpenFile(paths[i], os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(buf.Bytes()); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("captured %d base cases x %d shapes = %d entries, each run twice with identical output (%d small file, %d large file)",
		len(cases), len(goldenShapes), len(recs), len(small.Cases), len(large.Cases))
}
