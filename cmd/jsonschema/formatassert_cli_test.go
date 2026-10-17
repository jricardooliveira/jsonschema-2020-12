package main

import "testing"

// validation 7.2: --format-assertion makes a malformed format value invalid.
func TestFormatAssertionFlag(t *testing.T) {
	sch := writeTemp(t, `{"format":"ipv4"}`)
	cases := []struct {
		doc  string
		flag bool
		want int
	}{
		{`"192.168.01.1"`, true, 1},
		{`"192.168.1.1"`, true, 0},
		{`"192.168.01.1"`, false, 0},
		{`"192.168.1.1"`, false, 0},
		{`5`, true, 0},
	}
	for _, c := range cases {
		args := []string{"validate"}
		if c.flag {
			args = append(args, "--format-assertion")
		}
		code, _, e := runCLI(t, c.doc, append(args, sch)...)
		if code != c.want {
			t.Errorf("doc %s flag %v: code %d, want %d (%s)", c.doc, c.flag, code, c.want, e)
		}
	}
}
