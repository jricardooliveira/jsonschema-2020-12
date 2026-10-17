package jsonschema

import (
	"testing"

	"github.com/neoface/jsonschema/internal/perfdata"
)

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
