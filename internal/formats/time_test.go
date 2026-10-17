package formats

import "testing"

// RFC 3339 5.6 full-time.
func TestTime(t *testing.T) {
	table(t, "time", map[string]bool{
		"10:00:00Z": true, "10:00:00z": true, "10:00:00.123-08:00": true,
		"23:59:60Z": true, "00:29:60+00:30": true, "15:59:60-08:00": true,
		"10:00:60Z": false, "23:59:60+01:00": false,
		"10:00:00": false, "24:00:00Z": false, "10:60:00Z": false,
		"10:00:00.Z": false, "10:00:00+1:00": false, "": false,
	})
}
