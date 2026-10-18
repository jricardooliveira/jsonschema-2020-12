package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
)

// TestGolden builds the real CLI once, runs every golden case as its own
// process and compares stdout, stderr and exit code with the files captured
// from v1.1. It never writes a golden file.
func TestGolden(t *testing.T) {
	var recs []goldenRecord
	for _, f := range []string{goldenSmallFile, goldenLargeFile} {
		raw, err := os.ReadFile(filepath.Join(goldenRoot, f))
		if err != nil {
			t.Fatal(err)
		}
		var gf goldenFile
		if err := json.Unmarshal(raw, &gf); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		if gf.Header.Cases != len(gf.Cases) {
			t.Fatalf("%s: header says %d cases, file holds %d", f, gf.Header.Cases, len(gf.Cases))
		}
		recs = append(recs, gf.Cases...)
	}
	cases, err := goldenCases()
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != len(cases)*len(goldenShapes) {
		t.Fatalf("golden files hold %d entries, the case list needs %d", len(recs), len(cases)*len(goldenShapes))
	}
	byID := map[string]goldenCase{}
	for _, c := range cases {
		byID[c.ID] = c
	}
	bin, err := goldenBuild(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var failures []string
	fail := func(id, what string) {
		mu.Lock()
		failures = append(failures, id+": "+what)
		mu.Unlock()
	}
	ch := make(chan goldenRecord)
	var wg sync.WaitGroup
	for w := 0; w < runtime.NumCPU(); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for want := range ch {
				base, _, _ := strings.Cut(want.ID, "|")
				c, ok := byID[base]
				if !ok {
					fail(want.ID, "no such case")
					continue
				}
				got, err := goldenRun(bin, c, want.Shape)
				if err != nil {
					fail(want.ID, err.Error())
					continue
				}
				// The recorded call must be the call made now: a changed case
				// list or perfdata recipe fails outright instead of comparing
				// a different run against old output.
				if !slices.Equal(got.Args, want.Args) {
					fail(want.ID, "recorded args differ from the case list")
					continue
				}
				if got.Stdin != want.Stdin || got.StdinSize != want.StdinSize || got.StdinSHA256 != want.StdinSHA256 {
					fail(want.ID, "recorded stdin differs from the generated stdin")
					continue
				}
				if got.Exit != want.Exit {
					fail(want.ID, "exit code differs")
				}
				for _, s := range []struct {
					name      string
					got, want goldenOut
				}{{"stdout", got.Stdout, want.Stdout}, {"stderr", got.Stderr, want.Stderr}} {
					if s.got.Size != s.want.Size || s.got.SHA256 != s.want.SHA256 ||
						(s.want.Text != nil && (s.got.Text == nil || *s.got.Text != *s.want.Text)) {
						fail(want.ID, s.name+" differs")
					}
				}
			}
		}()
	}
	for _, r := range recs {
		ch <- r
	}
	close(ch)
	wg.Wait()
	if len(failures) > 0 {
		n := min(len(failures), 20)
		t.Fatalf("%d of %d golden entries differ (first %d): %s", len(failures), len(recs), n, strings.Join(failures[:n], "; "))
	}
}
