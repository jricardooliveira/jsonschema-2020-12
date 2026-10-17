package formats

func init() { Register("time", validTime) }

// validTime checks RFC 3339 5.6 full-time: HH:MM:SS[.frac] plus a required
// offset (Z, z or +/-HH:MM). Second 60 is valid only at 23:59:60 UTC.
func validTime(s string) bool {
	if len(s) < 9 || s[2] != ':' || s[5] != ':' || !digits(s[:2]) || !digits(s[3:5]) || !digits(s[6:8]) {
		return false
	}
	h, m, sec := num(s[:2]), num(s[3:5]), num(s[6:8])
	if h > 23 || m > 59 || sec > 60 {
		return false
	}
	rest := s[8:]
	if rest[0] == '.' {
		i := 1
		for i < len(rest) && rest[i] >= '0' && rest[i] <= '9' {
			i++
		}
		if i == 1 {
			return false
		}
		rest = rest[i:]
	}
	off := 0 // offset in minutes
	switch {
	case rest == "Z" || rest == "z":
	case len(rest) == 6 && (rest[0] == '+' || rest[0] == '-') && rest[3] == ':' && digits(rest[1:3]) && digits(rest[4:]):
		oh, om := num(rest[1:3]), num(rest[4:])
		if oh > 23 || om > 59 {
			return false
		}
		off = oh*60 + om
		if rest[0] == '-' {
			off = -off
		}
	default:
		return false
	}
	if sec == 60 {
		utc := ((h*60+m-off)%1440 + 1440) % 1440
		return utc == 23*60+59
	}
	return true
}
