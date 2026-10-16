package validation

import (
	"errors"
	"sort"
	"strconv"
	"strings"

	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "uniqueItems", Vocabulary: VocabularyURI, Compile: compileUniqueItems, Evaluate: evalUniqueItems})
}

func compileUniqueItems(_ *keywords.CompileContext, value any) (any, error) {
	b, ok := value.(bool)
	if !ok {
		return nil, errors.New("uniqueItems: must be a boolean")
	}
	return b, nil
}

func evalUniqueItems(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	arr, ok := instance.([]any)
	if !ok || !compiled.(bool) {
		return true, nil
	}
	groups, _ := findDuplicates(arr)
	for _, g := range groups {
		if err := ctx.Report("elements " + joinPositions(g) + " are equal"); err != nil {
			return false, err
		}
	}
	return len(groups) == 0, nil
}

// joinPositions renders "1, 4 and 9" from indices.
func joinPositions(g []int) string {
	var b strings.Builder
	for i, p := range g {
		switch {
		case i == 0:
		case i == len(g)-1:
			b.WriteString(" and ")
		default:
			b.WriteString(", ")
		}
		b.WriteString(strconv.Itoa(p))
	}
	return b.String()
}

// findDuplicates returns groups (size >= 2, ascending positions, ordered by
// first position) of deep-equal elements, and the number of full deep-equality
// comparisons made. Elements are bucketed by a full-content fingerprint first,
// so comparisons happen only inside a bucket (at most one per element).
func findDuplicates(arr []any) ([][]int, int) {
	type class struct {
		rep     any
		members []int
	}
	buckets := map[string][]*class{}
	var order []*class
	comparisons := 0
	var sb strings.Builder
	for i, el := range arr {
		sb.Reset()
		fingerprint(&sb, el)
		key := sb.String()
		var found *class
		for _, c := range buckets[key] {
			comparisons++
			if jsonvalue.Equal(c.rep, el) {
				found = c
				break
			}
		}
		if found == nil {
			found = &class{rep: el}
			buckets[key] = append(buckets[key], found)
			order = append(order, found)
		}
		found.members = append(found.members, i)
	}
	var groups [][]int
	for _, c := range order {
		if len(c.members) > 1 {
			groups = append(groups, c.members)
		}
	}
	sort.Slice(groups, func(a, b int) bool { return groups[a][0] < groups[b][0] })
	return groups, comparisons
}

// fingerprint writes an unambiguous canonical encoding of v: equal values give
// equal text. Strings are length-prefixed, object members sorted by key and
// numbers written as normalised sign, mantissa and exponent. Cost is linear in
// the size of v; recursion is bounded by the decoder's nesting limit.
func fingerprint(sb *strings.Builder, v any) {
	switch x := v.(type) {
	case nil:
		sb.WriteByte('n')
	case bool:
		if x {
			sb.WriteByte('t')
		} else {
			sb.WriteByte('f')
		}
	case string:
		sb.WriteByte('s')
		sb.WriteString(strconv.Itoa(len(x)))
		sb.WriteByte(':')
		sb.WriteString(x)
	case *jsonvalue.Number:
		c := x.Canonical()
		sb.WriteByte('d')
		sb.WriteString(strconv.Itoa(len(c)))
		sb.WriteByte(':')
		sb.WriteString(c)
	case []any:
		sb.WriteByte('[')
		sb.WriteString(strconv.Itoa(len(x)))
		sb.WriteByte(':')
		for _, e := range x {
			fingerprint(sb, e)
		}
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		sb.WriteByte('{')
		sb.WriteString(strconv.Itoa(len(x)))
		sb.WriteByte(':')
		for _, k := range keys {
			sb.WriteString(strconv.Itoa(len(k)))
			sb.WriteByte(':')
			sb.WriteString(k)
			fingerprint(sb, x[k])
		}
	}
}
