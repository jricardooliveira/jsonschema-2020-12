package formats

func init() { Register("hostname", validHostname) }

// validHostname checks RFC 1123 2.1: dot-separated labels of 1-63 letters,
// digits or hyphens, none starting or ending with a hyphen, at most 253
// bytes in all. A trailing dot fails.
func validHostname(s string) bool {
	if len(s) == 0 || len(s) > 253 {
		return false
	}
	start := 0
	for i := 0; i <= len(s); i++ {
		if i < len(s) && s[i] != '.' {
			c := s[i]
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-') {
				return false
			}
			continue
		}
		l := i - start
		if l == 0 || l > 63 || s[start] == '-' || s[i-1] == '-' {
			return false
		}
		start = i + 1
	}
	return true
}
