package formats

import "testing"

// ECMA-262 22.2.2.9 and the table of Unicode property names and values.
func TestECMAUnicodeProperty(t *testing.T) {
	good := []string{"Letter", "L", "Lu", "Cn", "LC", "gc=Lu", "General_Category=Letter",
		"Script=Greek", "sc=Grek", "Script_Extensions=Latin", "scx=Latn", "sc=Old_Italic",
		"ASCII", "Any", "Alphabetic", "Alpha", "Emoji", "White_Space"}
	bad := []string{"", "Lettter", "letter", "Script=Foo", "sc=", "Script", "Script=greek",
		"Foo=Greek", "ASCII=Yes", "Lu=Lu", "gc=Greek", "sc=Lu", "=", "L ", "sc=Grek=x"}
	for _, s := range good {
		if !validECMAUnicodeProperty(s) {
			t.Errorf("%q should be accepted", s)
		}
	}
	for _, s := range bad {
		if validECMAUnicodeProperty(s) {
			t.Errorf("%q should be rejected", s)
		}
	}
}
