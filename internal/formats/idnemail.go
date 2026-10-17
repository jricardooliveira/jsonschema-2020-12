package formats

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func init() { Register("idn-email", validIDNEmail) }

// validIDNEmail is a basic check only. Plain ASCII follows email. Otherwise:
// 254 bytes overall, a local part of 1-64 bytes without '@', space, control
// characters, or empty dot segments, and a domain that is an address literal
// or passes idn-hostname.
// NOT checked: RFC 6531 UTF8-non-ascii categories, normalization, quoted
// non-ASCII local parts (rejected if they hold a space), IDNA domain rules.
func validIDNEmail(s string) bool {
	if isASCII(s) {
		return validEmail(s)
	}
	if !utf8.ValidString(s) || len(s) > 254 {
		return false
	}
	i := strings.LastIndexByte(s, '@')
	if i < 1 || i > 64 {
		return false
	}
	local := s[:i]
	if strings.Contains(local, "@") || local[0] == '.' || local[len(local)-1] == '.' || strings.Contains(local, "..") {
		return false
	}
	for _, r := range local {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return false
		}
	}
	return emailDomain(s[i+1:], validIDNHostname)
}
