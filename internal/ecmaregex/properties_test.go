package ecmaregex

import "testing"

func TestResolveProperty(t *testing.T) {
	ok := map[string]string{
		"L": "L", "Letter": "L", "Lu": "Lu", "Uppercase_Letter": "Lu",
		"gc=Nd": "Nd", "General_Category=Decimal_Number": "Nd",
		"Script=Greek": "Greek", "sc=Latin": "Latin",
		"Punctuation": "P", "Other": "C",
	}
	for in, want := range ok {
		got, err := resolveProperty(in)
		if err != nil || got != want {
			t.Errorf("%q = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, in := range []string{"Greek", "Script=Nope", "scx=Greek", "Script_Extensions=Greek",
		"Alphabetic", "", "Cased_Letter", "foo=L"} {
		if got, err := resolveProperty(in); err == nil {
			t.Errorf("%q accepted as %q", in, got)
		}
	}
}

func TestEveryLongCategoryNameHasGoTable(t *testing.T) {
	for long := range longCategory {
		if _, err := category(long); err != nil {
			t.Errorf("%s: %v", long, err)
		}
	}
}
