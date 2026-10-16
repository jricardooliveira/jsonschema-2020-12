package ecmaregex

import (
	"errors"
	"regexp"
	"testing"
)

func TestTranslateMatches(t *testing.T) {
	cases := []struct {
		pat, in string
		want    bool
	}{
		{`^\d$`, "5", true}, {`^\d$`, "٣", false},
		{`^\D$`, "a", true}, {`^\D$`, "5", false}, {`^\D$`, "٣", true},
		{`^\w+$`, "aZ_9", true}, {`^\w$`, "é", false},
		{`^\W$`, "é", true}, {`^\W$`, "a", false},
		{`^\s$`, " ", true}, {`^\s$`, " ", true}, {`^\s$`, "a", false},
		{`^\S$`, "a", true}, {`^\S$`, " ", false},
		{`^.$`, "😀", true}, {`^.$`, "\n", false}, {`^.$`, "\r", false},
		{`^.$`, " ", false}, {`^.$`, " ", false}, {`^.$`, "ab", false},
		{`^A$`, "A", true},
		{`^😀$`, "😀", true},
		{`^\u{1F600}$`, "😀", true},
		{`^\cJ$`, "\n", true}, {`^\0$`, "\x00", true},
		{`^a\/b$`, "a/b", true},
		{`^(?<y>\d{4})$`, "2024", true}, {`^(?<y>\d{4})$`, "20x4", false},
		{`^\x41$`, "A", true}, {`^\t$`, "\t", true}, {`^\v$`, "\v", true},
		// class-internal shorthands
		{`^[\d]$`, "7", true}, {`^[\d]$`, "٣", false},
		{`^[\D]$`, "a", true}, {`^[\D]$`, "7", false},
		{`^[\w]$`, "é", false}, {`^[\W]$`, "-", true}, {`^[\W]$`, "a", false},
		{`^[\S]$`, "a", true}, {`^[\S]$`, " ", false},
		{`^[^\S]$`, " ", true}, {`^[^\S]$`, "a", false},
		{`^[\s\d]+$`, " 1", true},
		{`^[a-c]$`, "b", true}, {`^[a-c]$`, "d", false},
		{`^[A-C]$`, "B", true},
		{`^[😀-😂]$`, "😁", true},
		{`^[😀-😂]$`, "😃", false},
		{`^[\b]$`, "\b", true},
		{`^[.]$`, "x", false}, {`^[.]$`, ".", true},
		{`^[]$`, "a", false}, {`^[^]$`, "\n", true},
		{`^[a-]$`, "-", true}, {`^[-a]$`, "-", true},
		// accepted literals
		{`^\_$`, "_", true}, {`^\-$`, "-", true}, {`^[\-]$`, "-", true},
		{`^a{$`, "a{", true}, {`^a]$`, "a]", true}, {`^a}$`, "a}", true},
		{`^a{,5}$`, "a{,5}", true}, {`^a{2}$`, "aa", true}, {`^a{2,}$`, "aaa", true},
		{`^\.$`, ".", true}, {`^\.$`, "x", false},
		// RE2-only syntax does not get its RE2 meaning
		{`^[[:alpha:]]$`, "a", false}, {`^[[:alpha:]]$`, "a]", true},
		// properties
		{`^\p{Letter}+$`, "é", true}, {`^\p{L}$`, "1", false},
		{`^\P{L}$`, "1", true}, {`^\P{Letter}$`, "a", false},
		{`^\p{Script=Greek}$`, "α", true}, {`^\p{sc=Latin}$`, "α", false},
		{`^\p{gc=Lu}$`, "A", true}, {`^\p{General_Category=Uppercase_Letter}$`, "a", false},
		{`^[\p{Letter}\d]+$`, "é1", true},
		// unanchored and boundaries
		{`b`, "abc", true}, {`\bfoo\b`, "a foo b", true}, {`\bfoo\b`, "afoob", false},
	}
	for _, c := range cases {
		out, err := Translate(c.pat)
		if err != nil {
			t.Errorf("%q: %v", c.pat, err)
			continue
		}
		re, err := regexp.Compile(out)
		if err != nil {
			t.Errorf("%q -> %q does not compile: %v", c.pat, out, err)
			continue
		}
		if got := re.MatchString(c.in); got != c.want {
			t.Errorf("%q on %q = %v, want %v (re2 %q)", c.pat, c.in, got, c.want, out)
		}
	}
}

func TestTranslateRejects(t *testing.T) {
	for _, pat := range []string{
		`(?=a)`, `(?!a)`, `(?<=a)`, `(?<!a)`,
		`(a)\1`, `\k<n>`, `(?<n>a)\k<n>`,
		`\a`, `\z`, `\A`, `\Q.\E`, `\e`, `\é`,
		`(?i)a`, `(?i:a)`, `(?P<n>a)`, `(?<n!>a)`, `(?<1a>a)`, `(?<n`,
		`\ud800`, `\udc00`, `\ud83dA`, `\u{D800}`,
		`\u12`, `\u{110000}`, `\u{}`, `\u{12`, `\x4`, `\c1`, `\c`, `\00`,
		`\p{Script_Extensions=Greek}`, `\p{scx=Greek}`, `\p{Foo}`, `\p{Script=Nope}`,
		`\p{Greek}`, `\p{Unknown=L}`, `\p`, `\p{L`,
		`[a`, `\`, `[\d-z]`, `[z-a]`, `[\B]`, `[a-\d]`,
	} {
		out, err := Translate(pat)
		var e *Error
		if !errors.As(err, &e) {
			t.Errorf("%q: want *Error, got %v (out %q)", pat, err, out)
		}
	}
}
