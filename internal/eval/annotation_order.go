package eval

import (
	"slices"
	"strings"
)

var unescape = strings.NewReplacer("~1", "/", "~0", "~")

// tokens decodes a JSON Pointer into its segments (nil for the empty one).
func tokens(p string) []string {
	if p == "" {
		return nil
	}
	parts := strings.Split(p[1:], "/")
	for i, t := range parts {
		parts[i] = unescape.Replace(t)
	}
	return parts
}

// sortAnnotations returns a sorted copy of in.
func sortAnnotations(in []Annotation) []Annotation { return sortInPlace(slices.Clone(in)) }

// sortInPlace sorts in by instance location, then keyword location (decoded
// segments, a shorter prefix first), then absolute keyword location, keyword
// name and evaluation sequence. Segments are split once per annotation, into
// one shared slice, and an index permutation is sorted (fewer bytes moved).
func sortInPlace(in []Annotation) []Annotation {
	n := len(in)
	total := 0
	for _, a := range in {
		total += strings.Count(a.InstanceLocation, "/") + strings.Count(a.KeywordLocation, "/")
	}
	arena := make([]string, 0, total) // one segment per '/'
	off := make([]int, 2*n+1)         // segments of in[i]: instance off[2i]:off[2i+1], keyword off[2i+1]:off[2i+2]
	for i, a := range in {
		off[2*i] = len(arena)
		arena = appendTokens(arena, a.InstanceLocation)
		off[2*i+1] = len(arena)
		arena = appendTokens(arena, a.KeywordLocation)
	}
	off[2*n] = len(arena)
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	slices.SortStableFunc(idx, func(i, j int) int {
		if c := slices.Compare(arena[off[2*i]:off[2*i+1]], arena[off[2*j]:off[2*j+1]]); c != 0 {
			return c
		}
		if c := slices.Compare(arena[off[2*i+1]:off[2*i+2]], arena[off[2*j+1]:off[2*j+2]]); c != 0 {
			return c
		}
		x, y := &in[i], &in[j]
		if c := strings.Compare(x.AbsoluteKeywordLocation, y.AbsoluteKeywordLocation); c != 0 {
			return c
		}
		if c := strings.Compare(x.Keyword, y.Keyword); c != 0 {
			return c
		}
		return x.Seq - y.Seq
	})
	// apply the permutation in place, cycle by cycle: out[k] = in[idx[k]]
	for k := range idx {
		if idx[k] == k {
			continue
		}
		tmp, j := in[k], k
		for {
			next := idx[j]
			idx[j] = j
			if next == k {
				in[j] = tmp
				break
			}
			in[j] = in[next]
			j = next
		}
	}
	return in
}

// appendTokens appends the decoded segments of pointer p to dst.
func appendTokens(dst []string, p string) []string {
	if p == "" {
		return dst
	}
	s := p[1:]
	for {
		t, rest, more := strings.Cut(s, "/")
		dst = append(dst, unescape.Replace(t))
		if !more {
			return dst
		}
		s = rest
	}
}
