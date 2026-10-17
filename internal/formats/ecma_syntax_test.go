package formats

import (
	"strings"
	"testing"
)

// ECMA-262 22.2.1 (Pattern grammar, u flag).
func TestECMAPattern(t *testing.T) {
	good := []string{
		"", "a", "a|b|", "^a+$", "a*?", "a{2}", "a{2,}", "a{2,5}?", `\d+\.\w*`, `[a-z0-9_-]`, `[^\]]`, `[\d-]`, `[-a]`, `[a-]`,
		"(?=a)b", "(?!a)", "(?<=a)b", "(?<!a)b", "(?:a)", `(a)\1`, `\1(a)`, `(?<n>a)\k<n>`, `\k<n>(?<n>a)`,
		"(?<n>a)|(?<n>b)", `\/`, `A`, `\u{1F600}`, `😀`, `\x41`, `\cA`, `\0`, `\p{Letter}`, `\P{Script=Greek}`,
		`[\p{L}\d]`, `\b\B`, `[\b]`, `[\-]`, "(a(b)(?<x>c))\\3", "(?<$_é>a)",
		"a{", // placeholder replaced below
	}
	good = good[:len(good)-1]
	bad := []string{
		`\a`, "{", "a{", "a{,5}", "}", "]", "a{x}", "(", ")", "(a", "a)", "[", "[a", "[b-a]", `[\d-z]`, `[z-\d]`,
		"(?<n>a)(?<n>b)", "(?<n>a(?<n>b))", `\2(a)`, `\1`, `\k<n>`, `(?<n>a)\k<m>`, `\k`, "*", "+a", "a**", "a|*", "(*)",
		"(?=a)*", "(?<=a)+", "^*", `\b+`, "(?", "(?x)", "(?<>a)", "(?<1>a)", "(?<n", `\c`, `\c1`, `\u12`, `\u{110000}`, `\u{}`, `\x4`,
		`\00`, `\p{Lettter}`, `\p`, `\p{L`, `\-`, `\B+`, `[\B]`, `[\1]`, `[\k]`, "a{3,2}", "a{99999999999999999999,1}", "\xff", `\`,
	}
	for _, s := range good {
		if !validECMAPattern(s) {
			t.Errorf("%q should be valid", s)
		}
	}
	for _, s := range bad {
		if validECMAPattern(s) {
			t.Errorf("%q should be invalid", s)
		}
	}
	if !validECMAPattern("a{99999999999999999999,999999999999999999999}") {
		t.Error("huge ordered bounds should be valid")
	}
}

// A million unclosed groups or classes must be rejected without a crash.
func TestECMAPatternHostileNesting(t *testing.T) {
	for _, c := range []string{"(", "["} {
		if validECMAPattern(strings.Repeat(c, 1_000_000)) {
			t.Errorf("million %q accepted", c)
		}
	}
	if !validECMAPattern(strings.Repeat("(", 100000) + strings.Repeat(")", 100000)) {
		t.Error("deeply nested balanced groups should be valid")
	}
	if validECMAPattern(strings.Repeat("(?<a>", 50000) + strings.Repeat(")", 50000) + "(?<a>)") {
		t.Error("duplicate name accepted")
	}
}
