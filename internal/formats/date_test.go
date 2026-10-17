package formats

import "testing"

// RFC 3339 5.6 full-date.
func TestDate(t *testing.T) {
	table(t, "date", map[string]bool{
		"2024-02-29": true, "2000-02-29": true, "2023-02-28": true,
		"2023-02-29": false, "1900-02-29": false, "2024-13-01": false,
		"2024-04-31": false, "2024-00-10": false, "2024-1-01": false, "2024-01-1x": false, "": false,
	})
}
