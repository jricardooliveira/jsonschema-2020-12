package formats

func init() { Register("uuid", validUUID) }

// validUUID checks shape only (RFC 4122 3): 8-4-4-4-12 hex digits, either
// case. Version and variant digits are not examined.
func validUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if s[i] != '-' {
				return false
			}
		} else if !isHex(s[i]) {
			return false
		}
	}
	return true
}
