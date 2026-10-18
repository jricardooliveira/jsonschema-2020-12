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
