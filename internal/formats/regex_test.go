package formats

import (
	"errors"
	"testing"
)

func TestRegex(t *testing.T) {
	table(t, "regex", map[string]bool{
		"^a+$": true, `\d+`: true, "(?=a)": true, `(a)\1`: true, "(?<n>a)": true,
		`\p{Letter}`: true, `\p{L}`: true, `\p{Script=Greek}`: true, "": true,
		`\a`: false, "{": false, "(": false, "[b-a]": false, "(?<n>a)(?<n>b)": false,
		`\p{Lettter}`: false, `\p{Script=Foo}`: false,
	})
}

func TestTranslateContained(t *testing.T) {
	if translateContained(func(string) (string, error) { panic("boom") }, "a") {
		t.Error("panic must report false")
	}
	if translateContained(func(string) (string, error) { return "", errors.New("x") }, "a") {
		t.Error("error must report false")
	}
	check, _ := Lookup("regex")
	for _, p := range []string{"(?=a)", `(a)\1`, "(?<n>a)"} {
		if !check(p) {
			t.Errorf("%q should stay valid", p)
		}
	}
}
