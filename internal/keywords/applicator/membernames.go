package applicator

import "sort"

// memberNamesBuf is the caller's stack buffer for objects of up to 8 members.
type memberNamesBuf [8]string

// sortedMemberNames returns the object's member names in sort.Strings (byte)
// order, sorted once. Up to len(buf) names use buf (no heap allocation); more
// use a heap slice. The result aliases buf, so it is valid only while the
// caller's buffer is, and must not be retained past the keyword's evaluation.
func sortedMemberNames(obj map[string]any, buf *memberNamesBuf) []string {
	var keys []string
	if len(obj) <= len(buf) {
		keys = buf[:0]
	} else {
		keys = make([]string, 0, len(obj))
	}
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
