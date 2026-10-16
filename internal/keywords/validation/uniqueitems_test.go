package validation

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/jsonvalue"
)

// The testdata/validation/uniqueItems.json cases run through the root TestSpecCases.

func decodeArr(t *testing.T, s string) []any {
	t.Helper()
	v, err := jsonvalue.DecodeLimited([]byte(s), jsonvalue.NumberLimits{MaxExponent: 10_000_000})
	if err != nil {
		t.Fatal(err)
	}
	return v.([]any)
}

func TestFindDuplicatesGroups(t *testing.T) {
	arr := decodeArr(t, `[1, "a", 1.0, {"x":1,"y":[2]}, "a", {"y":[2.0],"x":1}, 1e0, "b"]`)
	got, _ := findDuplicates(arr)
	want := [][]int{{0, 2, 6}, {1, 4}, {3, 5}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("groups = %v, want %v", got, want)
	}
}

func TestJoinPositions(t *testing.T) {
	for in, want := range map[string][]int{"1 and 4": {1, 4}, "1, 4 and 9": {1, 4, 9}} {
		if got := joinPositions(want); got != in {
			t.Errorf("got %q want %q", got, in)
		}
	}
}

func TestHugeExponentsStayCheap(t *testing.T) {
	arr := decodeArr(t, `[1e1000000, 1e-1000000, 10e999999, 1e1000000]`)
	got, _ := findDuplicates(arr)
	if want := [][]int{{0, 2, 3}}; !reflect.DeepEqual(got, want) {
		t.Errorf("groups = %v, want %v", got, want)
	}
}

func TestLargeArrayNearLinear(t *testing.T) {
	const n = 100000
	var b strings.Builder
	b.WriteString("[")
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		switch i % 4 {
		case 0:
			fmt.Fprintf(&b, "%d", i)
		case 1:
			fmt.Fprintf(&b, `"s%d"`, i)
		case 2:
			fmt.Fprintf(&b, `{"a":%d,"b":"v"}`, i) // same shape, distinct content
		default:
			fmt.Fprintf(&b, `{"a":{"b":[%d]}}`, i)
		}
	}
	b.WriteString(",1,1.0]") // no other element is 1, so these two form the only group
	arr := decodeArr(t, b.String())
	groups, cmps := findDuplicates(arr)
	last := len(arr) - 1
	if len(groups) != 1 || !reflect.DeepEqual(groups[0], []int{last - 1, last}) {
		t.Fatalf("groups = %v", groups)
	}
	if cmps >= n*n/100 || cmps > 2*n {
		t.Errorf("comparisons = %d, not near-linear", cmps)
	}
}
