package jsonvalue

import "testing"

func TestEqualMixedRepresentations(t *testing.T) {
	eq := [][2]string{
		{"1", "1.0"}, {"1", "1e0"}, {"100", "1e2"}, {"0", "-0.0"},
		{"9223372036854775807", "9223372036854775807.0"},
		{"9223372036854775808", "92233720368547758080e-1"},
		{"12345678901234567890", "1234567890123456789e1"},
	}
	for _, p := range eq {
		a, b := num(t, p[0]), num(t, p[1])
		if !Equal(a, b) || !Equal(b, a) {
			t.Errorf("%s should equal %s", p[0], p[1])
		}
		if a.Canonical() != b.Canonical() {
			t.Errorf("%s and %s: canonical keys differ", p[0], p[1])
		}
	}
	ne := [][2]string{
		{"9223372036854775807", "9223372036854775808"},
		{"1", "-1"}, {"0.1", "0.10000000000000001"}, {"10", "1"},
		{"92233720368547758071", "9223372036854775807"},
	}
	for _, p := range ne {
		a, b := num(t, p[0]), num(t, p[1])
		if Equal(a, b) || Equal(b, a) || a.Canonical() == b.Canonical() {
			t.Errorf("%s should differ from %s", p[0], p[1])
		}
	}
}

func TestEqualNestedMixedNumbersNoAllocs(t *testing.T) {
	a := []any{num(t, "1"), map[string]any{"k": num(t, "2.50")}}
	b := []any{num(t, "1.0"), map[string]any{"k": num(t, "25e-1")}}
	if !Equal(a, b) {
		t.Fatal("nested mixed numbers should be equal")
	}
	if n := testing.AllocsPerRun(100, func() { Equal(a, b) }); n != 0 {
		t.Errorf("Equal allocated %v times", n)
	}
}
