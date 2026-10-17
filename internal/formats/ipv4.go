package formats

import "strings"

func init() { Register("ipv4", validIPv4) }

// validIPv4 checks a strict dotted-quad (RFC 2673 3.2): exactly four decimal
// parts, each 0-255, no leading zeros. Also used by the email domain literal.
func validIPv4(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}
	for _, p := range parts {
		if len(p) == 0 || len(p) > 3 || !digits(p) || (len(p) > 1 && p[0] == '0') || num(p) > 255 {
			return false
		}
	}
	return true
}
