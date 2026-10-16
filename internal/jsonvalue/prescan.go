package jsonvalue

import "unicode/utf8"

const (
	reasonUTF8      = "invalid UTF-8"
	reasonSurrogate = "unpaired surrogate escape"
)

func hex4(b []byte) (rune, bool) {
	if len(b) < 4 {
		return 0, false
	}
	var r rune
	for _, c := range b[:4] {
		switch {
		case c >= '0' && c <= '9':
			c -= '0'
		case c >= 'a' && c <= 'f':
			c = c - 'a' + 10
		case c >= 'A' && c <= 'F':
			c = c - 'A' + 10
		default:
			return 0, false
		}
		r = r<<4 | rune(c)
	}
	return r, true
}

// scanText is the only UTF-8 and surrogate check. It returns the byte offset
// and fixed reason of the first problem, or reason "" when the text is fine.
// onString (may be nil) gets each string literal's span (quotes included) and
// how many U+FFFD it holds, written literally or as an escape. Malformed
// escapes and control characters are left to the JSON decoder.
func scanText(data []byte, onString func(start, end, fffd int)) (int, string) {
	inStr, start, fffd := false, 0, 0
	for i := 0; i < len(data); {
		b := data[i]
		if !inStr {
			if b == '"' {
				inStr, start, fffd = true, i, 0
				i++
			} else if b >= utf8.RuneSelf {
				r, size := utf8.DecodeRune(data[i:])
				if r == utf8.RuneError && size == 1 {
					return i, reasonUTF8
				}
				i += size
			} else {
				i++
			}
			continue
		}
		switch {
		case b == '"':
			if onString != nil {
				onString(start, i+1, fffd)
			}
			inStr = false
			i++
		case b == '\\':
			if i+1 >= len(data) {
				return 0, ""
			}
			if data[i+1] != 'u' {
				i += 2
				continue
			}
			r, ok := hex4(data[i+2:])
			if !ok {
				i += 2 // decoder reports the bad escape
				continue
			}
			switch {
			case r == utf8.RuneError:
				fffd++
				i += 6
			case r >= 0xD800 && r < 0xDC00:
				lo, ok := rune(0), false
				if i+12 <= len(data) && data[i+6] == '\\' && data[i+7] == 'u' {
					lo, ok = hex4(data[i+8:])
				}
				if !ok || lo < 0xDC00 || lo > 0xDFFF {
					return i, reasonSurrogate
				}
				i += 12
			case r >= 0xDC00 && r <= 0xDFFF:
				return i, reasonSurrogate
			default:
				i += 6
			}
		case b >= utf8.RuneSelf:
			r, size := utf8.DecodeRune(data[i:])
			if r == utf8.RuneError && size == 1 {
				return i, reasonUTF8
			}
			if r == utf8.RuneError {
				fffd++
			}
			i += size
		default:
			i++
		}
	}
	return 0, ""
}
