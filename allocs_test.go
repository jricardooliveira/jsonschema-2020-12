package jsonschema

import (
	"runtime"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/perfdata"
)

// Allocation-count limits: the measured count (testing.AllocsPerRun, from raw
// bytes, see docs/benchmarks.md) plus about 10% headroom. One limit applies on
// every platform; if a newer Go exceeds one, re-measure on purpose and record why.
const (
	allocLimitCompile     = 1325    // measured: 1203
	allocLimitSmall       = 432     // measured: 392
	allocLimitMedium      = 331000  // measured: 300035
	allocLimitLarge       = 3301000 // measured: 3000069
	allocLimitErrors      = 266000  // measured: 240999
	allocLimitErrorsBasic = 272000  // measured: 247028
)

// Total-bytes limits: the runtime.MemStats TotalAlloc delta of one operation
// (every byte requested, even if freed at once), averaged over bytesRuns runs
// after a warm-up, plus about 10% headroom.
const (
	bytesRuns        = 3
	bytesLimitMedium = 13468000  // measured: 12243154
	bytesLimitLarge  = 137141000 // measured: 124673165
	bytesLimitErrors = 12770000  // measured: 11611584 (kept: +10% would be 12,772,743)
)

func checkBytes(t *testing.T, limit uint64, op func()) {
	t.Helper()
	skipUnderRace(t)
	op() // warm-up
	var a, b runtime.MemStats
	runtime.ReadMemStats(&a)
	for i := 0; i < bytesRuns; i++ {
		op()
	}
	runtime.ReadMemStats(&b)
	got := (b.TotalAlloc - a.TotalAlloc) / bytesRuns
	t.Logf("bytes %d, limit %d", got, limit)
	if got > limit {
		t.Errorf("bytes %d exceed limit %d", got, limit)
	}
}

func skipUnderRace(t *testing.T) {
	t.Helper()
	if raceEnabled {
		t.Skip("the race detector adds allocations the validator did not cause")
	}
}

func checkAllocs(t *testing.T, runs int, limit int, op func()) {
	t.Helper()
	skipUnderRace(t)
	op() // warm-up: builds the shared meta-schema set and any pools
	got := testing.AllocsPerRun(runs, op)
	t.Logf("allocs %.0f, limit %d", got, limit)
	if got > float64(limit) {
		t.Errorf("allocs %.0f exceed limit %d", got, limit)
	}
}

func allocSchema(t *testing.T) *Schema {
	t.Helper()
	src, err := perfdata.Fixture("order_schema.json")
	if err != nil {
		t.Fatal(err)
	}
	s, err := Compile(src)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func allocValidate(t *testing.T, doc []byte, wantValid, basic bool) func() {
	s := allocSchema(t)
	return func() {
		res, err := s.Validate(doc)
		if err != nil {
			t.Fatal(err)
		}
		if res.Valid() != wantValid {
			t.Fatalf("valid = %v, want %v", res.Valid(), wantValid)
		}
		if basic {
			if _, err := res.Output("basic"); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestAllocsCompileOrder(t *testing.T) {
	src, err := perfdata.Fixture("order_schema.json")
	if err != nil {
		t.Fatal(err)
	}
	checkAllocs(t, 20, allocLimitCompile, func() {
		if _, err := Compile(src); err != nil {
			t.Fatal(err)
		}
	})
}

func TestAllocsValidateSmall(t *testing.T) {
	doc, err := perfdata.Fixture("order_small.json")
	if err != nil {
		t.Fatal(err)
	}
	checkAllocs(t, 50, allocLimitSmall, allocValidate(t, doc, true, false))
}

func TestAllocsValidate0_5MB(t *testing.T) {
	op := allocValidate(t, perfdata.ValidOrder(perfdata.MediumItems), true, false)
	checkAllocs(t, 3, allocLimitMedium, op)
	checkBytes(t, bytesLimitMedium, op)
}

func TestAllocsValidate5_5MB(t *testing.T) {
	op := allocValidate(t, perfdata.ValidOrder(perfdata.LargeItems), true, false)
	checkAllocs(t, 2, allocLimitLarge, op)
	checkBytes(t, bytesLimitLarge, op)
}

func TestAllocsValidateErrors(t *testing.T) {
	op := allocValidate(t, perfdata.ErrorOrder(), false, false)
	checkAllocs(t, 3, allocLimitErrors, op)
	checkBytes(t, bytesLimitErrors, op)
}

func TestAllocsValidateErrorsBasic(t *testing.T) {
	checkAllocs(t, 3, allocLimitErrorsBasic, allocValidate(t, perfdata.ErrorOrder(), false, true))
}

// Story #206 (one child context per depth): limits = measured + ~10%.
const (
	ccAllocLimitSmall  = 560       // measured: 509
	ccBytesLimitSmall  = 25600     // measured: 23216
	ccBytesLimitLarge  = 216700000 // measured: 196951210
	ccBytesLimitErrors = 17100000  // measured: 15511248
)

func TestAllocsChildCtxValidateSmall(t *testing.T) {
	doc, err := perfdata.Fixture("order_small.json")
	if err != nil {
		t.Fatal(err)
	}
	op := allocValidate(t, doc, true, false)
	checkAllocs(t, 50, ccAllocLimitSmall, op)
	checkBytes(t, ccBytesLimitSmall, op)
}

func TestAllocsChildCtxValidate5_5MB(t *testing.T) {
	op := allocValidate(t, perfdata.ValidOrder(perfdata.LargeItems), true, false)
	checkBytes(t, ccBytesLimitLarge, op)
}

func TestAllocsChildCtxValidateErrors(t *testing.T) {
	op := allocValidate(t, perfdata.ErrorOrder(), false, false)
	checkBytes(t, ccBytesLimitErrors, op)
}

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

// Story 210 (multipleOf without big integers): allocation limits re-measured
// with this story, measured count plus about 10%. They are tighter than the
// limits in allocs_test.go, which stay as looser upper bounds. Helpers and
// race gating come from allocs_test.go.
const (
	allocLimitMultipleOfMedium      = 407000  // measured: 370045
	allocLimitMultipleOfLarge       = 4070000 // measured: 3700065
	allocLimitMultipleOfErrors      = 308000  // measured: 280003
	allocLimitMultipleOfErrorsBasic = 315000  // measured: 286036
)

func TestAllocsMultipleOfValidate0_5MB(t *testing.T) {
	op := allocValidate(t, perfdata.ValidOrder(perfdata.MediumItems), true, false)
	checkAllocs(t, 3, allocLimitMultipleOfMedium, op)
}

func TestAllocsMultipleOfValidate5_5MB(t *testing.T) {
	op := allocValidate(t, perfdata.ValidOrder(perfdata.LargeItems), true, false)
	checkAllocs(t, 2, allocLimitMultipleOfLarge, op)
}

func TestAllocsMultipleOfValidateErrors(t *testing.T) {
	checkAllocs(t, 3, allocLimitMultipleOfErrors, allocValidate(t, perfdata.ErrorOrder(), false, false))
}

func TestAllocsMultipleOfValidateErrorsBasic(t *testing.T) {
	checkAllocs(t, 3, allocLimitMultipleOfErrorsBasic, allocValidate(t, perfdata.ErrorOrder(), false, true))
}

// Story 207 limits (object keywords without wasted copies): measured value
// plus about 10% headroom. They are lower on purpose than the limits in
// allocs_test.go, which stay as looser upper bounds. The measured values
// include the other round-3 changes already in the tree.
const (
	objKeysAllocSmall       = 480     // measured: 435
	objKeysAllocMedium      = 374100  // measured: 340040
	objKeysAllocLarge       = 3740100 // measured: 3400062
	objKeysAllocErrors      = 265200  // measured: 241004
	objKeysAllocErrorsBasic = 271800  // measured: 247033

	objKeysBytesMedium = 16912600  // measured: 15375061
	objKeysBytesLarge  = 181446100 // measured: 164950997
	objKeysBytesErrors = 15514700  // measured: 14104240
)

func TestAllocsObjectKeysValidateSmall(t *testing.T) {
	doc, err := perfdata.Fixture("order_small.json")
	if err != nil {
		t.Fatal(err)
	}
	checkAllocs(t, 50, objKeysAllocSmall, allocValidate(t, doc, true, false))
}

func TestAllocsObjectKeysValidate0_5MB(t *testing.T) {
	op := allocValidate(t, perfdata.ValidOrder(perfdata.MediumItems), true, false)
	checkAllocs(t, 3, objKeysAllocMedium, op)
	checkBytes(t, objKeysBytesMedium, op)
}

func TestAllocsObjectKeysValidate5_5MB(t *testing.T) {
	op := allocValidate(t, perfdata.ValidOrder(perfdata.LargeItems), true, false)
	checkAllocs(t, 2, objKeysAllocLarge, op)
	checkBytes(t, objKeysBytesLarge, op)
}

func TestAllocsObjectKeysValidateErrors(t *testing.T) {
	op := allocValidate(t, perfdata.ErrorOrder(), false, false)
	checkAllocs(t, 3, objKeysAllocErrors, op)
	checkBytes(t, objKeysBytesErrors, op)
}

func TestAllocsObjectKeysValidateErrorsBasic(t *testing.T) {
	checkAllocs(t, 3, objKeysAllocErrorsBasic, allocValidate(t, perfdata.ErrorOrder(), false, true))
}
