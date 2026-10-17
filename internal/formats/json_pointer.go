package formats

import "unicode/utf8"

// RFC 6901 JSON Pointer, checked in one forward scan: "" or tokens each
// starting with '/', where '~' is followed only by '0' or '1'. Fully checked.
func init() { Register("json-pointer", validJSONPointer) }

func validJSONPointer(s string) bool {
	if !utf8.ValidString(s) {
		return false
	}
	if s == "" {
		return true
	}
	if s[0] != '/' {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] == '~' && (i+1 >= len(s) || (s[i+1] != '0' && s[i+1] != '1')) {
			return false
		}
	}
	return true
}
