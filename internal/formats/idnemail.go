package formats

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func init() { Register("idn-email", validIDNEmail) }

// validIDNEmail is a basic check only. Plain ASCII follows email. Otherwise:
// 254 bytes overall, an unquoted local part of 1-64 bytes whose ASCII
// characters are atext or '.' (RFC 6531 3.3), with no space, control
// character or empty dot segment, and a domain that is an address literal or
// passes idn-hostname.
// NOT checked: RFC 6531 UTF8-non-ascii categories, normalization, IDNA domain
// rules. A quoted non-ASCII local part is rejected (its '"' is not atext).
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
		if r < utf8.RuneSelf && r != '.' && !atext(byte(r)) || unicode.IsSpace(r) || unicode.IsControl(r) {
			return false
		}
	}
	return emailDomain(s[i+1:], validIDNHostname)
}
