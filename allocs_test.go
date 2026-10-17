package jsonschema

import (
	"testing"

	"github.com/neoface/jsonschema/internal/perfdata"
)

// Allocation-count limits: the measured count (testing.AllocsPerRun, from raw
// bytes, see docs/benchmarks.md) plus about 10% headroom. One limit applies on
// every platform; if a newer Go exceeds one, re-measure on purpose and record why.
const (
	allocLimitCompile     = 2730    // measured: 2481
	allocLimitSmall       = 864     // measured: 785
	allocLimitMedium      = 726000  // measured: 660073
	allocLimitLarge       = 7260000 // measured: 6600096
	allocLimitErrors      = 528000  // measured: 480018
	allocLimitErrorsBasic = 534700  // measured: 486042
)

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
	checkAllocs(t, 3, allocLimitMedium, allocValidate(t, perfdata.ValidOrder(perfdata.MediumItems), true, false))
}

func TestAllocsValidate5_5MB(t *testing.T) {
	checkAllocs(t, 2, allocLimitLarge, allocValidate(t, perfdata.ValidOrder(perfdata.LargeItems), true, false))
}

func TestAllocsValidateErrors(t *testing.T) {
	checkAllocs(t, 3, allocLimitErrors, allocValidate(t, perfdata.ErrorOrder(), false, false))
}

func TestAllocsValidateErrorsBasic(t *testing.T) {
	checkAllocs(t, 3, allocLimitErrorsBasic, allocValidate(t, perfdata.ErrorOrder(), false, true))
}
