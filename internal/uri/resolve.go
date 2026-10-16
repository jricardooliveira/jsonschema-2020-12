// Package uri implements RFC 3986 reference resolution, safe normalization
// and JSON Pointer fragment parsing. It never touches the network or disk.
package uri

import (
	"fmt"
	"regexp"
	"strings"
)

// RFC 3986 appendix B.
var uriRe = regexp.MustCompile(`^(?:([^:/?#]+):)?(?://([^/?#]*))?([^?#]*)(?:\?([^#]*))?(?:#(.*))?$`)

type parts struct {
	scheme, auth, path, query, frag       string
	hasScheme, hasAuth, hasQuery, hasFrag bool
}

func parse(s string) (parts, error) {
	if err := ValidatePercent(s); err != nil {
		return parts{}, err
	}
	m := uriRe.FindStringSubmatchIndex(s)
	if m == nil {
		return parts{}, fmt.Errorf("invalid URI reference %q", s)
	}
	g := func(i int) (string, bool) {
		if m[2*i] < 0 {
			return "", false
		}
		return s[m[2*i]:m[2*i+1]], true
	}
	var p parts
	p.scheme, p.hasScheme = g(1)
	p.auth, p.hasAuth = g(2)
	p.path, _ = g(3)
	p.query, p.hasQuery = g(4)
	p.frag, p.hasFrag = g(5)
	return p, nil
}

func (p parts) String() string {
	var b strings.Builder
	if p.hasScheme {
		b.WriteString(p.scheme + ":")
	}
	if p.hasAuth {
		b.WriteString("//" + p.auth)
	}
	b.WriteString(p.path)
	if p.hasQuery {
		b.WriteString("?" + p.query)
	}
	if p.hasFrag {
		b.WriteString("#" + p.frag)
	}
	return b.String()
}

// ValidatePercent reports an error when s holds a '%' not followed by two hex digits.
func ValidatePercent(s string) error {
	for i := 0; i < len(s); i++ {
		if s[i] == '%' {
			if i+2 >= len(s) || !isHex(s[i+1]) || !isHex(s[i+2]) {
				return fmt.Errorf("invalid percent-encoding in %q", s)
			}
			i += 2
		}
	}
	return nil
}

func isHex(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

func removeDotSegments(in string) string {
	var out []string
	segs := strings.Split(in, "/")
	for i, s := range segs {
		last := i == len(segs)-1
		switch s {
		case ".":
			if last {
				out = append(out, "")
			}
		case "..":
			if len(out) > 1 || (len(out) == 1 && out[0] != "") {
				out = out[:len(out)-1]
			}
			if last {
				out = append(out, "")
			}
		default:
			out = append(out, s)
		}
	}
	r := strings.Join(out, "/")
	if strings.HasPrefix(in, "/") && !strings.HasPrefix(r, "/") {
		r = "/" + r
	}
	return r
}

// Resolve resolves reference ref against base per RFC 3986 section 5.2.
func Resolve(base, ref string) (string, error) {
	r, err := parse(ref)
	if err != nil {
		return "", err
	}
	b, err := parse(base)
	if err != nil {
		return "", err
	}
	var t parts
	switch {
	case r.hasScheme:
		t = r
		t.path = removeDotSegments(r.path)
	case r.hasAuth:
		t = r
		t.scheme, t.hasScheme = b.scheme, b.hasScheme
		t.path = removeDotSegments(r.path)
	default:
		t.scheme, t.hasScheme = b.scheme, b.hasScheme
		t.auth, t.hasAuth = b.auth, b.hasAuth
		if r.path == "" {
			t.path = b.path
			if r.hasQuery {
				t.query, t.hasQuery = r.query, true
			} else {
				t.query, t.hasQuery = b.query, b.hasQuery
			}
		} else {
			if strings.HasPrefix(r.path, "/") {
				t.path = removeDotSegments(r.path)
			} else {
				m := r.path
				if b.hasAuth && b.path == "" {
					m = "/" + r.path
				} else if i := strings.LastIndex(b.path, "/"); i >= 0 {
					m = b.path[:i+1] + r.path
				}
				t.path = removeDotSegments(m)
			}
			t.query, t.hasQuery = r.query, r.hasQuery
		}
	}
	t.frag, t.hasFrag = r.frag, r.hasFrag
	return t.String(), nil
}

func isUnreserved(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' ||
		c == '-' || c == '.' || c == '_' || c == '~'
}

func unhex(c byte) byte {
	switch {
	case c >= 'a':
		return c - 'a' + 10
	case c >= 'A':
		return c - 'A' + 10
	}
	return c - '0'
}

// normPct uppercases escape digits and unescapes unreserved characters only
// (RFC 3986 6.2.2.1/6.2.2.2); reserved escapes such as %2F stay escaped.
// The input must already pass ValidatePercent.
func normPct(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '%' {
			v := unhex(s[i+1])<<4 | unhex(s[i+2])
			if isUnreserved(v) {
				b.WriteByte(v)
			} else {
				fmt.Fprintf(&b, "%%%02X", v)
			}
			i += 2
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// Normalize applies RFC 3986 6.2.2 syntax-based normalization: lowercase
// scheme and host, unreserved unescaping, uppercase escapes, dot-segment removal.
func Normalize(s string) (string, error) {
	p, err := parse(s)
	if err != nil {
		return "", err
	}
	p.scheme = strings.ToLower(p.scheme)
	if i := strings.LastIndex(p.auth, "@"); i >= 0 {
		p.auth = normPct(p.auth[:i+1]) + strings.ToLower(normPct(p.auth[i+1:]))
	} else {
		p.auth = strings.ToLower(normPct(p.auth))
	}
	p.path = removeDotSegments(normPct(p.path))
	p.query = normPct(p.query)
	p.frag = normPct(p.frag)
	return p.String(), nil
}
