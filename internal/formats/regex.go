package formats

import "github.com/neoface/jsonschema/internal/ecmaregex"

// ECMA-262 regular expression (u flag). The verdict comes from the strict
// syntax check in ecma_syntax.go alone. JavaScript-only constructs that RE2
// cannot run (lookahead, lookbehind, back-references, named groups) are
// judged by syntax only and count as valid.
// ponytail: the translation step below is advisory: it exists only to
// exercise the reused step, and its result can never turn a valid pattern
// into a failure. Partly checked: see ecma_syntax.go and ecma_unicode_props.go.
func init() { Register("regex", validRegex) }

func validRegex(s string) bool {
	if !validECMAPattern(s) {
		return false
	}
	if len(s) <= ecmaregex.DefaultMaxPatternSize {
		translateContained(ecmaregex.Translate, s)
	}
	return true
}

// translateContained runs tr on s, ignoring errors and recovering panics. It
// reports whether the pattern was translated. Callers must discard it for
// the verdict.
func translateContained(tr func(string) (string, error), s string) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	_, err := tr(s)
	return err == nil
}
