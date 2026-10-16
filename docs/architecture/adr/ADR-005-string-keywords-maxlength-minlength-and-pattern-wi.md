# ADR #5: String keywords maxLength, minLength and pattern, with a reusable ECMA-262-to-RE2 translation step

- Date: 2026-10-16
- Story: #5
- Status: proposed (pending architecture sign-off)

## Decision
New package internal/ecmaregex contains translate.go, a single-pass scanner over runes that tracks whether it is inside a character class, and properties.go, which maps long General_Category names to short ones and validates script names against unicode.Scripts and unicode.Categories. It exposes only Translate(pattern string) (string, error), which returns a typed error describing the unsupported construct. The keyword files under internal/keywords/validation/ register name, vocabulary URI (https://json-schema.org/draft/2020-12/vocab/validation), compile function and evaluate function through internal/keywords/registry.go, so no central list is edited. pattern compiles *regexp.Regexp once at Compile time and stores it in the immutable compiled keyword. *regexp.Regexp is safe for concurrent use. The pattern-size limit is read from the compile options carried by the compile context. The only edit to the crowded file cmd/jsonschema/main.go is a single flag registration line passed to the existing options. internal/eval/eval.go, the other crowded file, is not touched.

## Affected modules
- internal/ecmaregex/translate.go
- internal/ecmaregex/properties.go
- internal/ecmaregex/translate_test.go
- internal/ecmaregex/properties_test.go
- internal/keywords/validation/maxlength.go
- internal/keywords/validation/maxlength_test.go
- internal/keywords/validation/minlength.go
- internal/keywords/validation/minlength_test.go
- internal/keywords/validation/pattern.go
- internal/keywords/validation/pattern_test.go
- testdata/validation/maxLength.json
- testdata/validation/minLength.json
- testdata/validation/pattern.json
- options.go
- internal/compile/limits.go
- cmd/jsonschema/main.go
- patternsize_test.go
- cmd/jsonschema/patternsize_test.go
- docs/regex-deviations.md

## Data / API impact
- DB impact: no
- API impact: yes: adds a pattern-size library option and CLI flag
- Migration needed: no

## Constraints for implementation
- Do not modify internal/jsonvalue/decode.go. Assume all strings are valid Unicode (story 179).
- Standard library only (regexp, unicode, unicode/utf8, strings). No go.mod require lines.
- Translate the ECMA rules for the dot and the shorthands both outside and inside character classes. Inside a class, a negated shorthand such as \D must become explicit ranges, because RE2 classes cannot nest negations.
- Combine two \u escapes forming a surrogate pair into one code point. A lone surrogate escape is a schema error.
- Never pass RE2-only syntax through with its RE2 meaning. Escape RE2 metacharacters that are literal in ECMA, for example '[' inside a class, so [[:alpha:]] is not read as a POSIX class.
- Match unanchored with MatchString on the Go string, so '.' and lengths both work per code point.
- Check the pattern-size limit before translation and compilation.
- Error and failure messages never include instance values. The pattern text in the message is allowed.
- Each keyword's tests go in its own _test.go and its own testdata file. Tests read limit defaults from code.
- Touch cmd/jsonschema/main.go with one flag line only.

## Risks
- The translator's character-class handling is subtle. Negated shorthands inside negated classes, ranges whose ends are escapes, and surrogate pairs inside classes all need dedicated tests.
- Go's General_Category and script tables may differ slightly from ECMA's list of property names, so the supported list must be documented precisely.
- main.go is crowded (3 stories), so this story serialises with the others that edit it.
- The option and flag names and their file locations depend on story 179's accepted plan, which this proposal cannot see.

## Breaking changes
- none

## Sensitivity
- security
