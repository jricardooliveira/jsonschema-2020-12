package formats

func init() { Register("date", validDate) }

func digits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return len(s) > 0
}

func num(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		n = n*10 + int(s[i]-'0')
	}
	return n
}

// validDate checks RFC 3339 5.6 full-date: YYYY-MM-DD, a real calendar day.
func validDate(s string) bool {
	if len(s) != 10 || s[4] != '-' || s[7] != '-' || !digits(s[:4]) || !digits(s[5:7]) || !digits(s[8:]) {
		return false
	}
	y, m, d := num(s[:4]), num(s[5:7]), num(s[8:])
	if m < 1 || m > 12 || d < 1 {
		return false
	}
	days := [...]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	max := days[m-1]
	if m == 2 && (y%4 == 0 && (y%100 != 0 || y%400 == 0)) {
		max = 29
	}
	return d <= max
}
