package eval

import (
	"reflect"
	"sync"
	"testing"
)

func TestEvaluatedItemsRecordAndRead(t *testing.T) {
	r := NewEvaluatedItems()
	r.RecordPrefixItems("", "/prefixItems", 1, false)
	r.RecordItems("", "/items")
	r.RecordContains("", "/contains", []int{0, 2}, false)
	r.RecordContains("/a", "/contains", []int{0, 1}, true)

	got := r.At("")
	if a := got["/prefixItems"]; a.All || a.Largest != 1 {
		t.Errorf("prefixItems = %+v", a)
	}
	if !got["/items"].All {
		t.Error("items should be true")
	}
	if !reflect.DeepEqual(got["/contains"].Indices, []int{0, 2}) {
		t.Errorf("contains = %+v", got["/contains"])
	}
	if a := r.At("/a")["/contains"]; !a.All || a.Indices != nil {
		t.Errorf("all-match contains = %+v", a)
	}
	if len(r.At("/none")) != 0 {
		t.Error("unrecorded location should be empty")
	}
}

func TestEvaluatedItemsRollbackDropsFailedBranch(t *testing.T) {
	r := NewEvaluatedItems()
	r.RecordItems("", "/items")
	m := r.Mark()
	r.RecordPrefixItems("", "/allOf/0/prefixItems", 0, true)
	r.Rollback(m)
	got := r.At("")
	if _, ok := got["/allOf/0/prefixItems"]; ok || !got["/items"].All {
		t.Errorf("rollback wrong: %+v", got)
	}
}

func TestEvaluatedItemsContainsCopiesIndices(t *testing.T) {
	r := NewEvaluatedItems()
	idx := []int{1}
	r.RecordContains("", "/contains", idx, false)
	idx[0] = 9
	if r.At("")["/contains"].Indices[0] != 1 {
		t.Error("indices alias caller slice")
	}
}

func TestEvaluatedItemsSeparateRecordsRace(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := NewEvaluatedItems() // one record per call, never shared
			r.RecordItems("", "/items")
			_ = r.At("")
		}()
	}
	wg.Wait()
}
