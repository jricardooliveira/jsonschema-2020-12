package compile

import "strings"

const hexDigits = "0123456789ABCDEF"

// encodeFragment percent-encodes a JSON Pointer for use as a URI fragment
// (RFC 3986 section 3.5): pchar, '/' and '?' stay as they are; every other
// byte, including '%' and each UTF-8 byte of non-ASCII text, becomes %XX.
// '~' is unreserved, so the ~0/~1 escapes are never touched.
func encodeFragment(p string) string {
	var b strings.Builder
	for i := 0; i < len(p); i++ {
		c := p[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9',
			strings.IndexByte("-._~!$&'()*+,;=:@/?", c) >= 0:
			b.WriteByte(c)
		default:
			b.WriteByte('%')
			b.WriteByte(hexDigits[c>>4])
			b.WriteByte(hexDigits[c&15])
		}
	}
	return b.String()
}
