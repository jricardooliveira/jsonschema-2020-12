package formats

// draft-bhutton-relative-json-pointer-00: non-negative-integer (no leading
// zero), optional index manipulation ("+N" or "-N"), then "#" or a JSON
// pointer. Fully checked; nothing is resolved.
func init() { Register("relative-json-pointer", validRelativeJSONPointer) }

// digitRun returns the length of the leading run of ASCII digits in s.
func digitRun(s string) int {
	n := 0
	for n < len(s) && s[n] >= '0' && s[n] <= '9' {
		n++
	}
	return n
}

// validNonNegInt is "0" or a digit run without a leading zero.
func validNonNegInt(d string) bool { return d != "" && (d == "0" || d[0] != '0') }

func validRelativeJSONPointer(s string) bool {
	n := digitRun(s)
	if !validNonNegInt(s[:n]) {
		return false
	}
	s = s[n:]
	if s != "" && (s[0] == '+' || s[0] == '-') {
		m := digitRun(s[1:])
		if !validNonNegInt(s[1 : 1+m]) {
			return false
		}
		s = s[1+m:]
	}
	return s == "#" || validJSONPointer(s)
}
