package formats

import (
	"strings"
	"unicode/utf8"
)

// maxTemplatePrefix is the largest ":N" prefix modifier (RFC 6570 2.4.1).
const maxTemplatePrefix = 9999

// RFC 6570 URI Template: the spelling is checked in one forward scan, the
// template is never expanded.
// ponytail: partly checked: only the level 4 grammar is checked. The
// reserved-for-future operators (= , ! @ |) are rejected, and nothing
// beyond the grammar (for example a sensible variable list) is judged.
func init() { Register("uri-template", validURITemplate) }

func validURITemplate(s string) bool {
	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case c == '{':
			j := strings.IndexByte(s[i+1:], '}')
			if j < 0 || !validTemplateExpr(s[i+1:i+1+j]) {
				return false
			}
			i += j + 2
		case c == '%':
			if i+2 >= len(s) || !isHex(s[i+1]) || !isHex(s[i+2]) {
				return false
			}
			i += 3
		case c < utf8.RuneSelf:
			if c < 0x21 || strings.IndexByte("\"'<>\\^`|}", c) >= 0 {
				return false
			}
			i++
		default:
			r, n := utf8.DecodeRuneInString(s[i:])
			if !isUcsChar(r) && !isIPrivate(r) {
				return false
			}
			i += n
		}
	}
	return true
}

// validTemplateExpr checks the inside of "{...}".
func validTemplateExpr(e string) bool {
	if e != "" && strings.IndexByte("+#./;?&", e[0]) >= 0 {
		e = e[1:]
	}
	for _, spec := range strings.Split(e, ",") {
		if strings.HasSuffix(spec, "*") {
			spec = spec[:len(spec)-1]
		} else if i := strings.IndexByte(spec, ':'); i >= 0 {
			if !validTemplatePrefix(spec[i+1:]) {
				return false
			}
			spec = spec[:i]
		}
		if !validTemplateVarname(spec) {
			return false
		}
	}
	return true
}

// validTemplatePrefix is 1..maxTemplatePrefix with no leading zero, read
// with at most 4 digits so it cannot overflow.
func validTemplatePrefix(n string) bool {
	if n == "" || len(n) > 4 || n[0] == '0' {
		return false
	}
	v := 0
	for i := 0; i < len(n); i++ {
		if n[i] < '0' || n[i] > '9' {
			return false
		}
		v = v*10 + int(n[i]-'0')
	}
	return v <= maxTemplatePrefix
}

// validTemplateVarname is varchar *(["."] varchar).
func validTemplateVarname(v string) bool {
	if v == "" {
		return false
	}
	prevDot := true // a leading dot is not allowed
	for i := 0; i < len(v); {
		c := v[i]
		switch {
		case c == '.':
			if prevDot {
				return false
			}
			prevDot = true
			i++
		case c == '%':
			if i+2 >= len(v) || !isHex(v[i+1]) || !isHex(v[i+2]) {
				return false
			}
			prevDot = false
			i += 3
		case isAlpha(c) || (c >= '0' && c <= '9') || c == '_':
			prevDot = false
			i++
		default:
			return false
		}
	}
	return !prevDot
}
