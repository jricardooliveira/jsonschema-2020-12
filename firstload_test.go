package jsonschema

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/perfdata"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/registry"
)

// TestFirstLoadBuildsBuiltinIndexOnce checks, in a fresh process, that the
// first public Compile (AddBuiltins plus the meta-schema cache) builds the
// built-in index exactly once. The count is exact, so race mode needs no
// separate limit.
func TestFirstLoadBuildsBuiltinIndexOnce(t *testing.T) {
	if _, child := freshProcess(t); !child {
		return
	}
	if _, err := Compile([]byte(`{"type":"string"}`)); err != nil {
		t.Fatal(err)
	}
	if n := registry.BuiltinBuilds(); n != 1 {
		t.Fatalf("built-in index builds = %d, want 1", n)
	}
}

const freshChildEnv = "JSONSCHEMA_FRESH_CHILD"

// freshProcess runs the calling test alone in a fresh process (a re-exec of
// the test binary with an anchored -test.run). In the parent it runs the
// child, fails t if the child fails, and returns the child's combined output
// with child=false: the caller must then return. In the child it returns
// child=true and the caller runs its body, so nothing else compiled a schema
// first. Callers must measure at the very start of their body.
func freshProcess(t *testing.T) (out string, child bool) {
	t.Helper()
	if os.Getenv(freshChildEnv) == t.Name() {
		return "", true
	}
	cmd := exec.Command(os.Args[0], "-test.run=^"+t.Name()+"$", "-test.v")
	cmd.Env = append(os.Environ(), freshChildEnv+"="+t.Name())
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("fresh child failed: %v\n%s", err, b)
	}
	return string(b), false
}

var firstLoadTiming = flag.Bool("firstload", false, "measure first-Compile time, bytes and allocations (TestFirstLoadTiming)")

const firstLoadRuns = 10

// TestFirstLoadTiming starts about 10 fresh test programs, measures only the
// first Compile of the order schema in each, and logs the medians. It asserts
// no timing. Run: go test -run TestFirstLoadTiming -firstload -v .
func TestFirstLoadTiming(t *testing.T) {
	if os.Getenv(freshChildEnv) == t.Name() { // child: measure at the very start
		src, err := perfdata.Fixture("order_schema.json")
		if err != nil {
			t.Fatal(err)
		}
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		start := time.Now()
		if _, err := Compile(src); err != nil {
			t.Fatal(err)
		}
		d := time.Since(start)
		runtime.ReadMemStats(&after)
		fmt.Printf("FIRSTLOAD %d %d %d\n", d.Nanoseconds(), after.TotalAlloc-before.TotalAlloc, after.Mallocs-before.Mallocs)
		return
	}
	if !*firstLoadTiming {
		t.Skip("opt-in: pass -firstload to measure the first Compile")
	}
	var ns, bytes, allocs []uint64
	for i := 0; i < firstLoadRuns; i++ {
		out, _ := freshProcess(t)
		sc := bufio.NewScanner(strings.NewReader(out))
		found := false
		for sc.Scan() {
			var a, b, c uint64
			if n, _ := fmt.Sscanf(sc.Text(), "FIRSTLOAD %d %d %d", &a, &b, &c); n == 3 {
				ns, bytes, allocs = append(ns, a), append(bytes, b), append(allocs, c)
				found = true
			}
		}
		if !found {
			t.Fatalf("run %d: no FIRSTLOAD line in child output:\n%s", i, out)
		}
	}
	t.Logf("first Compile over %d fresh processes: median %d ns, %d B, %d allocs",
		firstLoadRuns, median(ns), median(bytes), median(allocs))
}

func median(v []uint64) uint64 {
	s := slices.Clone(v)
	slices.Sort(s)
	return s[len(s)/2]
}
