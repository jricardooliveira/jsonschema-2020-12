package formats

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func init() { Register("idn-hostname", validIDNHostname) }

// validIDNHostname is a basic check only (no IDNA tables, stdlib only).
// Plain ASCII follows hostname. Otherwise it checks the overall (253) and
// per-label (63) length in code points, no empty label, no trailing dot, no
// hyphen at a label edge, no space or control character, and ASCII characters
// limited to letters, digits and '-'.
// NOT checked: IDNA2008 code point categories (PVALID/CONTEXTJ/CONTEXTO),
// normalization, bidi rules, "--" in positions 3-4, Unicode punctuation,
// and the ideographic full stops.
func validIDNHostname(s string) bool {
	if isASCII(s) {
		return validHostname(s)
	}
	if !utf8.ValidString(s) || utf8.RuneCountInString(s) > 253 {
		return false
	}
	for _, l := range strings.Split(s, ".") {
		n := utf8.RuneCountInString(l)
		if n == 0 || n > 63 || l[0] == '-' || l[len(l)-1] == '-' {
			return false
		}
		for _, r := range l {
			if r < utf8.RuneSelf {
				if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-') {
					return false
				}
			} else if unicode.IsSpace(r) || unicode.IsControl(r) {
				return false
			}
		}
	}
	return true
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}
