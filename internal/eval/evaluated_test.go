package eval

import (
	"reflect"
	"sync"
	"testing"
)

func TestRecordPropertiesReadBack(t *testing.T) {
	r := NewEvaluatedItems()
	in := []string{"a", "b"}
	r.RecordProperties("/x", "/properties", in)
	r.RecordProperties("/y", "/properties", []string{"c"})
	in[0] = "mutated"
	got := r.PropertiesAt("/x")
	if !reflect.DeepEqual(got, map[string][]string{"/properties": {"a", "b"}}) {
		t.Fatalf("got %v", got)
	}
	if len(r.At("/x")) != 0 {
		t.Fatal("item view must skip property entries")
	}
}

func TestRecordPropertiesRollback(t *testing.T) {
	r := NewEvaluatedItems()
	r.RecordProperties("", "/properties", []string{"a"})
	m := r.Mark()
	r.RecordProperties("", "/patternProperties", []string{"b"})
	r.Rollback(m)
	got := r.PropertiesAt("")
	if len(got) != 1 || got["/properties"] == nil {
		t.Fatalf("got %v", got)
	}
}

func TestRecordPropertiesIndependentRecords(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := NewEvaluatedItems()
			r.RecordProperties("", "/properties", []string{"a"})
			if n := len(r.PropertiesAt("")["/properties"]); n != 1 {
				t.Error("shared state")
			}
		}()
	}
	wg.Wait()
}
