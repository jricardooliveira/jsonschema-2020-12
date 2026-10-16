package output

import "testing"

func TestFlagExact(t *testing.T) {
	for valid, want := range map[bool]string{true: `{"valid":true}`, false: `{"valid":false}`} {
		b, err := Render("flag", valid, nil)
		if err != nil || string(b) != want {
			t.Errorf("got %s %v", b, err)
		}
	}
}
