package formats

import "strings"

func init() { Register("email", validEmail) }

// atext reports an RFC 5322 3.2.3 atom character.
func atext(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' ||
		strings.IndexByte("!#$%&'*+-/=?^_`{|}~", c) >= 0
}

// dotAtom checks atext runs separated by single dots.
func dotAtom(s string) bool {
	if s == "" {
		return false
	}
	prev := byte('.')
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '.' {
			if prev == '.' {
				return false
			}
		} else if !atext(c) {
			return false
		}
		prev = c
	}
	return prev != '.'
}

// quoted checks an RFC 5321 4.1.2 Quoted-string (with its quotes).
func quoted(s string) bool {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return false
	}
	in := s[1 : len(s)-1]
	for i := 0; i < len(in); i++ {
		c := in[i]
		if c == '\\' {
			i++
			if i >= len(in) || in[i] < 32 || in[i] > 126 {
				return false
			}
		} else if c < 32 || c > 126 || c == '"' {
			return false
		}
	}
	return true
}

// emailDomain checks an RFC 5321 domain: host(name) or a [address literal].
func emailDomain(d string, host func(string) bool) bool {
	if len(d) > 2 && d[0] == '[' && d[len(d)-1] == ']' {
		in := d[1 : len(d)-1]
		// The "IPv6:" tag is an ABNF literal, so case-insensitive (RFC 5234 2.3).
		if len(in) >= 5 && strings.EqualFold(in[:5], "IPv6:") {
			return validIPv6(in[5:])
		}
		return validIPv4(in)
	}
	return host(d)
}

// splitEmail splits at the '@' ending the local part, which may be quoted.
func splitEmail(s string) (local, domain string, ok bool) {
	if len(s) > 0 && s[0] == '"' {
		for i := 1; i < len(s); i++ {
			switch s[i] {
			case '\\':
				i++
			case '"':
				if i+1 < len(s) && s[i+1] == '@' {
					return s[:i+1], s[i+2:], true
				}
				return "", "", false
			}
		}
		return "", "", false
	}
	return strings.Cut(s, "@")
}

// validEmail checks an RFC 5321 Mailbox: dot-atom or quoted local part of at
// most 64 bytes, '@', a hostname or address-literal domain, 254 bytes at most.
func validEmail(s string) bool {
	if len(s) > 254 {
		return false
	}
	local, domain, ok := splitEmail(s)
	if !ok || len(local) > 64 || !(dotAtom(local) || quoted(local)) {
		return false
	}
	return emailDomain(domain, validHostname)
}
