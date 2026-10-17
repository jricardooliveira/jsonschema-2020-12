package formats

import (
	"strings"
	"unicode/utf8"
)

// validRFC3986 reports whether s is an RFC 3986 URI (needScheme) or
// URI-reference; with iri it follows RFC 3987 and also admits ucschar (and
// iprivate in the query). It is a single forward scan: each component is
// split off by index and checked with scanChars; nothing recurses over the
// input and no regexp or net/url is used.
// ponytail: partly checked: the host is only checked as reg-name or
// IP-literal syntax (no IPv4 dotted-quad rule, no IDN/bidi rules of RFC 3987
// 4.2), and port digits are not range-limited.
func validRFC3986(s string, iri, needScheme bool) bool {
	if !utf8.ValidString(s) {
		return false
	}
	if i := strings.IndexByte(s, '#'); i >= 0 {
		if !scanChars(s[i+1:], ":@/?", iri, false) {
			return false
		}
		s = s[:i]
	}
	if i := strings.IndexByte(s, '?'); i >= 0 {
		if !scanChars(s[i+1:], ":@/?", iri, true) {
			return false
		}
		s = s[:i]
	}
	hasScheme := false
	if i := strings.IndexAny(s, ":/"); i >= 0 && s[i] == ':' {
		if !validScheme(s[:i]) {
			return false
		}
		hasScheme, s = true, s[i+1:]
	}
	if needScheme && !hasScheme {
		return false
	}
	if strings.HasPrefix(s, "//") {
		s = s[2:]
		auth, path := s, ""
		if i := strings.IndexByte(s, '/'); i >= 0 {
			auth, path = s[:i], s[i:]
		}
		return validAuthority(auth, iri) && scanChars(path, ":@/", iri, false)
	}
	if !hasScheme && !strings.HasPrefix(s, "/") {
		// path-noscheme: the first segment must not contain a colon.
		first := s
		if i := strings.IndexByte(s, '/'); i >= 0 {
			first = s[:i]
		}
		if strings.IndexByte(first, ':') >= 0 {
			return false
		}
	}
	return scanChars(s, ":@/", iri, false)
}

func validScheme(s string) bool {
	if s == "" || !isAlpha(s[0]) {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !isAlpha(c) && !(c >= '0' && c <= '9') && c != '+' && c != '-' && c != '.' {
			return false
		}
	}
	return true
}

func isAlpha(c byte) bool { return c|0x20 >= 'a' && c|0x20 <= 'z' }

func validAuthority(a string, iri bool) bool {
	if i := strings.IndexByte(a, '@'); i >= 0 {
		if !scanChars(a[:i], ":", iri, false) {
			return false
		}
		a = a[i+1:]
	}
	var port string
	if strings.HasPrefix(a, "[") {
		end := strings.IndexByte(a, ']')
		if end < 0 || !validIPLiteral(a[1:end]) {
			return false
		}
		a = a[end+1:]
		if a == "" {
			return true
		}
		if a[0] != ':' {
			return false
		}
		port = a[1:]
	} else {
		host := a
		if i := strings.IndexByte(a, ':'); i >= 0 {
			host, port = a[:i], a[i+1:]
		}
		if !scanChars(host, "", iri, false) {
			return false
		}
	}
	for i := 0; i < len(port); i++ {
		if port[i] < '0' || port[i] > '9' {
			return false
		}
	}
	return true
}

// validIPLiteral checks the inside of "[...]": IPv6address or IPvFuture.
func validIPLiteral(s string) bool {
	if len(s) > 1 && (s[0] == 'v' || s[0] == 'V') {
		i := 1
		for i < len(s) && isHex(s[i]) {
			i++
		}
		if i == 1 || i+1 >= len(s) || s[i] != '.' {
			return false
		}
		for j := i + 1; j < len(s); j++ {
			if c := s[j]; !isUnreserved(c) && !isSubDelim(c) && c != ':' {
				return false
			}
		}
		return true
	}
	return validIPv6(s)
}

func isUnreserved(c byte) bool {
	return isAlpha(c) || (c >= '0' && c <= '9') || c == '-' || c == '.' || c == '_' || c == '~'
}

func isSubDelim(c byte) bool { return strings.IndexByte("!$&'()*+,;=", c) >= 0 }

// scanChars checks that s is made of unreserved, sub-delims, well-formed
// pct-encodings, the bytes in extra and, in IRI mode, ucschar (plus iprivate
// when priv is set).
func scanChars(s, extra string, iri, priv bool) bool {
	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case c == '%':
			if i+2 >= len(s) || !isHex(s[i+1]) || !isHex(s[i+2]) {
				return false
			}
			i += 3
		case c < utf8.RuneSelf:
			if !isUnreserved(c) && !isSubDelim(c) && strings.IndexByte(extra, c) < 0 {
				return false
			}
			i++
		default:
			r, n := utf8.DecodeRuneInString(s[i:])
			if !iri || !(isUcsChar(r) || (priv && isIPrivate(r))) {
				return false
			}
			i += n
		}
	}
	return true
}

// isUcsChar is RFC 3987 ucschar.
func isUcsChar(r rune) bool {
	switch {
	case r >= 0xA0 && r <= 0xD7FF, r >= 0xF900 && r <= 0xFDCF, r >= 0xFDF0 && r <= 0xFFEF:
		return true
	case r >= 0x10000 && r <= 0xDFFFD:
		return r&0xFFFF < 0xFFFE
	case r >= 0xE1000 && r <= 0xEFFFD:
		return true
	}
	return false
}

// isIPrivate is RFC 3987 iprivate.
func isIPrivate(r rune) bool {
	return (r >= 0xE000 && r <= 0xF8FF) || (r >= 0xF0000 && r <= 0xFFFFD) || (r >= 0x100000 && r <= 0x10FFFD)
}
