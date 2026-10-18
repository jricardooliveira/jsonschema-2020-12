package jsonschema

import (
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/perfdata"
)

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
