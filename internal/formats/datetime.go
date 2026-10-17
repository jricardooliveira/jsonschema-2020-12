package formats

func init() { Register("date-time", validDateTime) }

// validDateTime checks RFC 3339 5.6 date-time: full-date, 'T' or 't', full-time.
func validDateTime(s string) bool {
	return len(s) > 11 && (s[10] == 'T' || s[10] == 't') && validDate(s[:10]) && validTime(s[11:])
}
