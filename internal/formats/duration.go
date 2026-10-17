package formats

func init() { Register("duration", validDuration) }

// consumeUnit reads "<digits><unit>"; it returns the text after it and the
// units after the one used. The first unit of a run may be any of units; a
// later one must be units[0], the next adjacent unit, since the ABNF chains
// them (dur-year = 1*DIGIT "Y" [dur-month], dur-hour = 1*DIGIT "H"
// [dur-minute], ...).
func consumeUnit(s, units string, first bool) (string, string, bool) {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 || i == len(s) {
		return s, units, false
	}
	for j := 0; j < len(units) && (first || j == 0); j++ {
		if units[j] == s[i] {
			return s[i+1:], units[j+1:], true
		}
	}
	return s, units, false
}

// validDuration checks RFC 3339 appendix A: P nW alone, or P, a run of
// adjacent date units from Y, M, D (P1Y1D fails), then optionally T and a
// non-empty run of adjacent time units from H, M, S (PT1H1S fails).
func validDuration(s string) bool {
	if len(s) < 2 || s[0] != 'P' {
		return false
	}
	s = s[1:]
	if rest, _, ok := consumeUnit(s, "W", true); ok && rest == "" {
		return true
	}
	n := 0
	for units := "YMD"; len(s) > 0 && s[0] != 'T'; n++ {
		var ok bool
		if s, units, ok = consumeUnit(s, units, n == 0); !ok {
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
	for units, k := "HMS", 0; len(s) > 0; k++ {
		var ok bool
		if s, units, ok = consumeUnit(s, units, k == 0); !ok {
			return false
		}
	}
	return true
}
