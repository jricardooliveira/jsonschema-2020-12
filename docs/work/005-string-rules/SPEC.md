# SPEC #5: String keywords maxLength, minLength and pattern, with a reusable ECMA-262-to-RE2 translation step

- Date: 2026-10-16
- Type: feature
- Verdict: pass
- Status: proposed (pending Checkpoint 1 sign-off)

## Problem

The validator cannot yet check string length or string patterns. The 2020-12 spec defines 'pattern' with ECMA-262 regular expressions, but Go's regexp package uses RE2 syntax, so a naive pass-through gives wrong answers. Examples: \d matches non-ASCII digits in Go's Unicode classes, '.' matches U+2028, and \uXXXX and \p{Letter} are rejected. There is also no protection against oversized patterns in hostile schemas.

## Why this matters

maxLength, minLength and pattern are mandatory 2020-12 validation keywords listed in the brief's 'Mandatory scope'. The translation step is also needed later by patternProperties (story 183) and the regex format (story 196). The operator has fixed the counting rules, the translation rules, the error rules and the pattern-size limit in the clarifications.

## Acceptance criteria

- maxLength and minLength each live in their own file under internal/keywords/validation/. Each registers itself from init() through the keyword registry, and no central keyword list or switch is edited.
- maxLength and minLength count Unicode code points of the already-valid string. A string of N multi-byte characters, emoji included, has length N. They apply only to string instances: every other type passes. A non-negative-integer value such as 2.0 is accepted as the integer 2 using the exact number model.
- 'pattern' lives in its own file and registers through the keyword registry. It applies only to string instances, and it matches unanchored, as ECMA-262 does: the pattern may match anywhere in the string unless it uses ^ or $.
- A reusable translation step in its own package (internal/ecmaregex) turns an ECMA-262 pattern into an RE2 pattern with Unicode-mode semantics, as follows. \d and \D mean ASCII 0-9. \w and \W mean ASCII [0-9A-Za-z_]. \s and \S use ECMA's whitespace and line-terminator set. '.' excludes \n, \r, U+2028 and U+2029. All of these also work inside character classes. \uXXXX, surrogate pairs written as two \u escapes, \u{...}, \cX, \0 (not followed by a digit), \/ and (?<name>...) are translated. A pattern counts characters the way maxLength does: ^.$ matches a single emoji.
- Unicode property escapes accept both short and long General_Category names, for example \p{L} and \p{Letter}, and \P{...} as well. Script=/sc= names that Go has tables for are also accepted. Any property name Go has no table for, including Script_Extensions/scx, is a schema error.
- Identity escapes with a clear meaning are accepted as literals: for example \_, \- and other escaped syntax or punctuation characters. A lone { or ] is also accepted as a literal. An escape with no clear meaning, such as \a, is a schema error. Constructs that are valid in RE2 but not in ECMA-262, such as \z, \A, \Q...\E, (?i) flags and [[:alpha:]] POSIX classes, never silently get their RE2 meaning.
- A pattern that the translation step or RE2 cannot handle is a schema error with exit code 2, never 'invalid'. This covers lookahead and lookbehind, backreferences (\1, \k<name>), lone surrogate escapes, unknown escapes and anything RE2 refuses to compile, such as a repeat count over 1000. The error message names the problem and the schema location.
- 'pattern' has a size limit with a generous default. It is named in the same style as story 179's number-size option and flag, and the caller can raise it through a library option and a CLI flag. A pattern over the limit is a schema error with exit code 2. Tests read the default from the code instead of repeating the number.
- A failing 'pattern' reports a message that shows the pattern from the schema, for example 'does not match pattern ^[a-z]+$', and never the instance value. maxLength and minLength messages never contain the instance value either.
- Spec-derived positive and negative tests cite validation spec sections 6.3.1-6.3.3 and are stored as testdata/validation/maxLength.json, minLength.json and pattern.json, run by the shared table-driven runner. They cover multi-byte characters, emoji, non-string instances and translated patterns. Unit tests for the translation step cover every translated construct and every rejected construct.
- docs/regex-deviations.md lists what the translation step translates, what it accepts as literals, what is a schema error, the supported Unicode property names, and every remaining difference between ECMA-262 and RE2.
- Compiled patterns are part of the immutable compiled Schema. Evaluation keeps no shared mutable state, and the tests pass under go test -race.

## Non-goals

- Changing internal/jsonvalue/decode.go or any JSON decoding. Rejecting invalid UTF-8 and lone surrogate escapes in documents is story 179's job, and this story relies on strings already being valid Unicode.
- patternProperties (story 183) and the regex format (story 196). They will reuse internal/ecmaregex later, but nothing is added for them now.
- Supporting lookaround, backreferences or any non-RE2 regex engine.
- The general limits on nesting depth, recursion and input size (story 193).
- Output-format rendering changes (stories 180 and 194). This story only supplies the keyword error messages.
- Regex flags beyond the implicit Unicode mode. JSON Schema patterns carry no flags.

## Tasks

- **T1** — ECMA-262-to-RE2 translation package
  - purpose: Write a reusable translator, internal/ecmaregex.Translate(pattern) (string, error), that implements the operator's translation, literal and error rules. Unicode property name mapping goes in its own file.
  - scope: internal/ecmaregex/translate.go, internal/ecmaregex/properties.go, internal/ecmaregex/translate_test.go, internal/ecmaregex/properties_test.go
  - done when: Unit tests pass for each translated construct (\d \D \w \W \s \S, '.', \uXXXX, surrogate pairs, \u{...}, \cX, \0, \/, named groups, class-internal shorthands), each accepted literal (\_, \-, lone { and ]), each rejected construct (lookaround, backreferences, \a, lone surrogates, RE2-only syntax, unknown properties, Script_Extensions), and ^.$ matching one emoji after compilation.
  - delivers: AC4, AC5, AC6, AC7
- **T2** — maxLength and minLength keywords
  - purpose: Add two self-registering keyword files that count code points with utf8.RuneCountInString and ignore non-string instances. Messages never contain the instance.
  - scope: internal/keywords/validation/maxlength.go, internal/keywords/validation/maxlength_test.go, internal/keywords/validation/minlength.go, internal/keywords/validation/minlength_test.go, testdata/validation/maxLength.json, testdata/validation/minLength.json
  - done when: Spec-derived testdata cases pass through the shared runner, covering ASCII, multi-byte, emoji, boundary values, 2.0 as a limit and non-string instances. A test asserts that messages omit the instance value.
  - delivers: AC1, AC2, AC9, AC10
- **T3** — pattern keyword
  - purpose: Add a self-registering pattern keyword. At compile time it translates the pattern with internal/ecmaregex and compiles it with regexp, turning failures into schema errors. At evaluation it matches string instances unanchored and reports 'does not match pattern <pattern>'.
  - scope: internal/keywords/validation/pattern.go, internal/keywords/validation/pattern_test.go, testdata/validation/pattern.json
  - done when: Spec-derived positive and negative cases pass, including unanchored matching, translated shorthands, \p{Letter}, emoji with ^.$ and non-string instances. A lookahead or backreference pattern returns a schema error from Compile. The CLI exits with code 2 for such a schema. The tests pass under -race.
  - depends on: T1
  - delivers: AC3, AC7, AC9, AC10, AC12
- **T4** — Pattern size limit option and CLI flag
  - purpose: Add a pattern-size limit, named in the style of story 179's number-size option. It is a library option fixed at compile time, and the CLI exposes it through a flag wired with one line in main.go. Patterns over the limit are schema errors.
  - scope: options.go, internal/compile/limits.go, cmd/jsonschema/main.go, patternsize_test.go, cmd/jsonschema/patternsize_test.go
  - done when: A library test shows that a pattern one character over the default limit (read from the code) is a schema error and that raising the option accepts it. A CLI contract test shows exit code 2 by default and exit code 0 or 1 once the flag raises the limit.
  - depends on: T3
  - delivers: AC8
- **T5** — Regex deviation documentation
  - purpose: Document the translation rules, the accepted literals, the errors, the supported property names, the pattern-size limit and the remaining ECMA-262 versus RE2 differences.
  - scope: docs/regex-deviations.md
  - done when: The document lists every construct covered by the T1 tests and every remaining difference.
  - depends on: T1, T4
  - delivers: AC11
