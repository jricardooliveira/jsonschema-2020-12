package formats

import "strings"

func init() { Register("ipv6", validIPv6) }

func isHex(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

// ipv6Groups counts the 16-bit groups in a ':'-separated run. An embedded
// IPv4 tail (allowed only when allow4 and last) counts as two. ok is false on
// any malformed group.
func ipv6Groups(part string, allow4 bool) (int, bool) {
	if part == "" {
		return 0, true
	}
	groups := strings.Split(part, ":")
	n := 0
	for i, g := range groups {
		if allow4 && i == len(groups)-1 && strings.Contains(g, ".") {
			if !validIPv4(g) {
				return 0, false
			}
			n += 2
			continue
		}
		if len(g) == 0 || len(g) > 4 {
			return 0, false
		}
		for j := 0; j < len(g); j++ {
			if !isHex(g[j]) {
				return 0, false
			}
		}
		n++
	}
	return n, true
}

// validIPv6 checks the RFC 4291 2.2 text forms, including '::' compression
// and an embedded IPv4 tail. Zone suffixes ('%') fail. Also used by email.
func validIPv6(s string) bool {
	head, tail, comp := strings.Cut(s, "::")
	if !comp {
		n, ok := ipv6Groups(s, true)
		return ok && n == 8
	}
	if strings.Contains(tail, "::") {
		return false
	}
	hn, ok := ipv6Groups(head, false)
	if !ok {
		return false
	}
	tn, ok := ipv6Groups(tail, true)
	return ok && hn+tn <= 7
}
