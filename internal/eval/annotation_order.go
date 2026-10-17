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

// keyed pairs an annotation with its pointer segments, computed once.
type keyed struct {
	a        Annotation
	ind, kwd []string
}

// sortAnnotations sorts by instance location, then keyword location (decoded
// segments, a shorter prefix first), then absolute keyword location, keyword
// name and evaluation sequence. Segments are split once per annotation.
func sortAnnotations(in []Annotation) []Annotation {
	ks := make([]keyed, len(in))
	for i, a := range in {
		ks[i] = keyed{a, tokens(a.InstanceLocation), tokens(a.KeywordLocation)}
	}
	slices.SortStableFunc(ks, func(x, y keyed) int {
		if c := slices.Compare(x.ind, y.ind); c != 0 {
			return c
		}
		if c := slices.Compare(x.kwd, y.kwd); c != 0 {
			return c
		}
		if c := strings.Compare(x.a.AbsoluteKeywordLocation, y.a.AbsoluteKeywordLocation); c != 0 {
			return c
		}
		if c := strings.Compare(x.a.Keyword, y.a.Keyword); c != 0 {
			return c
		}
		return x.a.Seq - y.a.Seq
	})
	out := make([]Annotation, len(ks))
	for i, k := range ks {
		out[i] = k.a
	}
	return out
}
