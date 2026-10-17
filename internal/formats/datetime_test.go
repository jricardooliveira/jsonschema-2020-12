package formats

import "testing"

// table checks the registered format name against expected results.
func table(t *testing.T, name string, cases map[string]bool) {
	t.Helper()
	check, ok := Lookup(name)
	if !ok {
		t.Fatalf("%s not registered", name)
	}
	for in, want := range cases {
		if got := check(in); got != want {
			t.Errorf("%s(%q) = %v, want %v", name, in, got, want)
		}
	}
}

// RFC 3339 5.6.
func TestDateTime(t *testing.T) {
	table(t, "date-time", map[string]bool{
		"2024-01-01T10:00:00Z":        true,
		"2024-01-01t10:00:00z":        true,
		"2024-01-01T10:00:00.5+01:00": true,
		"2016-12-31T23:59:60Z":        true,
		"2016-12-31T18:59:60-05:00":   true,
		"2016-12-31T20:59:60-05:00":   false,
		"2016-12-31T23:59:60+01:00":   false,
		"2024-01-01T10:00:00":         false,
		"2024-01-01 10:00:00Z":        false,
		"2023-02-29T10:00:00Z":        false,
		"":                            false,
	})
}
