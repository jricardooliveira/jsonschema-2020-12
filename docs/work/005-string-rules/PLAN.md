# PLAN #5: String keywords maxLength, minLength and pattern, with a reusable ECMA-262-to-RE2 translation step

- Date: 2026-10-16
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-005-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/ecmaregex/translate.go`
- `internal/ecmaregex/properties.go`
- `internal/ecmaregex/translate_test.go`
- `internal/ecmaregex/properties_test.go`
- `internal/keywords/validation/maxlength.go`
- `internal/keywords/validation/maxlength_test.go`
- `internal/keywords/validation/minlength.go`
- `internal/keywords/validation/minlength_test.go`
- `testdata/validation/maxLength.json`
- `testdata/validation/minLength.json`
- `internal/keywords/validation/pattern.go`
- `internal/keywords/validation/pattern_test.go`
- `testdata/validation/pattern.json`
- `options.go`
- `internal/compile/limits.go`
- `cmd/jsonschema/main.go`
- `patternsize_test.go`
- `cmd/jsonschema/patternsize_test.go`
- `docs/regex-deviations.md`

Modules the architecture expects to touch:

- `internal/ecmaregex/translate.go`
- `internal/ecmaregex/properties.go`
- `internal/ecmaregex/translate_test.go`
- `internal/ecmaregex/properties_test.go`
- `internal/keywords/validation/maxlength.go`
- `internal/keywords/validation/maxlength_test.go`
- `internal/keywords/validation/minlength.go`
- `internal/keywords/validation/minlength_test.go`
- `internal/keywords/validation/pattern.go`
- `internal/keywords/validation/pattern_test.go`
- `testdata/validation/maxLength.json`
- `testdata/validation/minLength.json`
- `testdata/validation/pattern.json`
- `options.go`
- `internal/compile/limits.go`
- `cmd/jsonschema/main.go`
- `patternsize_test.go`
- `cmd/jsonschema/patternsize_test.go`
- `docs/regex-deviations.md`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — ECMA-262-to-RE2 translation package
2. **T2** — maxLength and minLength keywords
3. **T3** — pattern keyword  _(after T1)_
4. **T4** — Pattern size limit option and CLI flag  _(after T3)_
5. **T5** — Regex deviation documentation  _(after T1, T4)_

## Risks

- The translator's character-class handling is subtle. Negated shorthands inside negated classes, ranges whose ends are escapes, and surrogate pairs inside classes all need dedicated tests.
- Go's General_Category and script tables may differ slightly from ECMA's list of property names, so the supported list must be documented precisely.
- main.go is crowded (3 stories), so this story serialises with the others that edit it.
- The option and flag names and their file locations depend on story 179's accepted plan, which this proposal cannot see.

Implementation constraints the coder must honor:

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

- DB impact: no · API impact: yes: adds a pattern-size library option and CLI flag · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: Unit tests pass for each translated construct (\d \D \w \W \s \S, '.', \uXXXX, surrogate pairs, \u{...}, \cX, \0, \/, named groups, class-internal shorthands), each accepted literal (\_, \-, lone { and ]), each rejected construct (lookaround, backreferences, \a, lone surrogates, RE2-only syntax, unknown properties, Script_Extensions), and ^.$ matching one emoji after compilation.
- **T2**: Spec-derived testdata cases pass through the shared runner, covering ASCII, multi-byte, emoji, boundary values, 2.0 as a limit and non-string instances. A test asserts that messages omit the instance value.
- **T3**: Spec-derived positive and negative cases pass, including unanchored matching, translated shorthands, \p{Letter}, emoji with ^.$ and non-string instances. A lookahead or backreference pattern returns a schema error from Compile. The CLI exits with code 2 for such a schema. The tests pass under -race.
- **T4**: A library test shows that a pattern one character over the default limit (read from the code) is a schema error and that raising the option accepts it. A CLI contract test shows exit code 2 by default and exit code 0 or 1 once the flag raises the limit.
- **T5**: The document lists every construct covered by the T1 tests and every remaining difference.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
