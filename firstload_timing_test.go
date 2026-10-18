package jsonschema

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/perfdata"
)

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
