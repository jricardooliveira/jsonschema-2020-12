package formats

func init() { Register("duration", validDuration) }

// consumeUnit reads "<digits><unit>" where unit is one of units; it returns
// the text after it and the units still allowed (those after the one used).
func consumeUnit(s, units string) (string, string, bool) {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 || i == len(s) {
		return s, units, false
	}
	for j := 0; j < len(units); j++ {
		if units[j] == s[i] {
			return s[i+1:], units[j+1:], true
		}
	}
	return s, units, false
}

// validDuration checks RFC 3339 appendix A: P nW alone, or
// P [nY][nM][nD] [T [nH][nM][nS]] in that order, with at least one component
// and a non-empty T part when T is present.
func validDuration(s string) bool {
	if len(s) < 2 || s[0] != 'P' {
		return false
	}
	s = s[1:]
	if rest, _, ok := consumeUnit(s, "W"); ok && rest == "" {
		return true
	}
	n := 0
	for units := "YMD"; len(s) > 0 && s[0] != 'T'; n++ {
		var ok bool
		if s, units, ok = consumeUnit(s, units); !ok {
			return false
		}
	}
	if len(s) == 0 {
		return n > 0
	}
	s = s[1:] // 'T'
	if len(s) == 0 {
		return false
	}
	for units := "HMS"; len(s) > 0; {
		var ok bool
		if s, units, ok = consumeUnit(s, units); !ok {
			return false
		}
	}
	return true
}
