package jsonschema

import (
	"runtime"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/perfdata"
)

// allocLimitFirstLoad bounds the allocations of the first Compile of the order
// schema in a fresh process (built-in index built once, meta-schema compiled).
// Measured: 8653 allocations; limit = measured + about 10%.
const allocLimitFirstLoad = 9520

// TestAllocsFirstLoadOrder counts the mallocs of the very first Compile in a
// fresh process. Allocation counts differ under -race, so it is skipped there.
func TestAllocsFirstLoadOrder(t *testing.T) {
	skipUnderRace(t)
	if _, child := freshProcess(t); !child {
		return
	}
	src, err := perfdata.Fixture("order_schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	if _, err := Compile(src); err != nil {
		t.Fatal(err)
	}
	runtime.ReadMemStats(&after)
	got := after.Mallocs - before.Mallocs
	t.Logf("first-load allocs = %d (limit %d)", got, allocLimitFirstLoad)
	if got > allocLimitFirstLoad {
		t.Fatalf("first Compile allocated %d times, limit %d", got, allocLimitFirstLoad)
	}
}
