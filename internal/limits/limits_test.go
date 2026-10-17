package limits

import "testing"

func TestResolve(t *testing.T) {
	const def = 7
	if v, err := Resolve("F", 0, def); err != nil || v != def {
		t.Errorf("zero: %d %v", v, err)
	}
	if v, err := Resolve("F", def+1, def); err != nil || v != def+1 {
		t.Errorf("set: %d %v", v, err)
	}
	if _, err := Resolve("F", -1, def); err == nil || err.Error() != "options: F must not be negative" {
		t.Errorf("negative: %v", err)
	}
}
