package eval

import (
	"slices"
	"strings"
	"testing"
)

// oldSort is the previous pointer-splitting comparator, kept as the oracle.
func oldSort(in []Annotation) []Annotation {
	out := slices.Clone(in)
	cmp := func(a, b string) int { return slices.Compare(tokens(a), tokens(b)) }
	slices.SortStableFunc(out, func(a, b Annotation) int {
		if c := cmp(a.InstanceLocation, b.InstanceLocation); c != 0 {
			return c
		}
		if c := cmp(a.KeywordLocation, b.KeywordLocation); c != 0 {
			return c
		}
		if c := strings.Compare(a.AbsoluteKeywordLocation, b.AbsoluteKeywordLocation); c != 0 {
			return c
		}
		if c := strings.Compare(a.Keyword, b.Keyword); c != 0 {
			return c
		}
		return a.Seq - b.Seq
	})
	return out
}

func TestSortAnnotationsMatchesPointerSplittingOrder(t *testing.T) {
	locs := []string{"", "/a", "/a/b", "/a~1b", "/a~0b", "/a~01", "/a/", "/", "/0", "/1", "/10", "/2", "/a%20b", "/ab"}
	var in []Annotation
	seq := 0
	for _, i := range locs {
		for _, k := range locs {
			seq++
			in = append(in, Annotation{InstanceLocation: i, KeywordLocation: k, Keyword: "kw" + k[:min(1, len(k))], Seq: seq % 7})
		}
	}
	// reverse so the sort has work to do
	slices.Reverse(in)
	if got, want := sortAnnotations(in), oldSort(in); !slices.Equal(got, want) {
		t.Fatal("new order differs from the old pointer-splitting order")
	}
	if got, want := SortAnnotations(in), oldSort(in); !slices.Equal(got, want) {
		t.Fatal("SortAnnotations differs from the old order")
	}
}
