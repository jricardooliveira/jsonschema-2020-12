package eval

import "testing"

func TestSkippedRecordRecordsNothing(t *testing.T) {
	for _, skip := range []bool{false, true} {
		r := newEvaluatedItems(10)
		r.skip = skip
		r.RecordProperties("", "/properties", []string{"a"})
		r.RecordItems("/x", "/items")
		r.RecordPrefixItems("/x", "/prefixItems", 0, false)
		r.RecordContains("/x", "/contains", []int{1}, false)
		got := len(r.PropertiesAt("")) + len(r.At("/x"))
		if want := map[bool]int{false: 4, true: 0}[skip]; got != want {
			t.Errorf("skip=%v: recorded %d entries, want %d", skip, got, want)
		}
		// Step charging is independent of the switch.
		if err := r.Charge(10); err != nil {
			t.Errorf("skip=%v: charge: %v", skip, err)
		}
		if err := r.Step(); err == nil {
			t.Errorf("skip=%v: step cap not enforced", skip)
		}
	}
}
