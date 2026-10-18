package jsonschema

import (
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/perfdata"
)

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
