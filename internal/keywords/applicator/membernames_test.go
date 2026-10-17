package applicator

import (
	"fmt"
	"reflect"
	"testing"
)

var byteOrderNames = []string{"B", "a", "é", "z", "", "ab", "a", "A", "aa", "Z", "e", "ÿ", "日本", "a1", "a10", "a2"}

func objOf(n int) map[string]any {
	m := map[string]any{}
	for i := 0; i < n; i++ {
		if i < len(byteOrderNames) {
			m[byteOrderNames[i]] = i
		} else {
			m[fmt.Sprintf("k%03d", i)] = i
		}
	}
	return m
}

func TestSortedMemberNamesMatchesSortStrings(t *testing.T) {
	for _, n := range []int{0, 1, 2, 8, 9, 15, 100} {
		obj := objOf(n)
		var buf memberNamesBuf
		got := sortedMemberNames(obj, &buf)
		want := sortedKeys(obj)
		if len(got) == 0 && len(want) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("n=%d: got %q want %q", n, got, want)
		}
	}
}

func TestSortedMemberNamesByteOrder(t *testing.T) {
	var buf memberNamesBuf
	got := sortedMemberNames(map[string]any{"a": 1, "B": 1, "é": 1, "z": 1, "": 1, "ab": 1}, &buf)
	want := []string{"", "B", "a", "ab", "z", "é"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestSortedMemberNamesNoAllocUpTo8(t *testing.T) {
	obj := objOf(8)
	n := 0
	allocs := testing.AllocsPerRun(100, func() {
		var buf memberNamesBuf
		n += len(sortedMemberNames(obj, &buf))
	})
	if allocs != 0 {
		t.Errorf("allocs = %v, want 0", allocs)
	}
	_ = n
}
