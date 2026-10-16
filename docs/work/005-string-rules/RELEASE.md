# Release notes — #5: String keywords maxLength, minLength and pattern, with a reusable ECMA-262-to-RE2 translation step

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #5
- Status: approved by the operator at Checkpoint 3 (run #138, 2026-10-16 20:11 UTC)

## What changed

Adds three validation keywords for JSON strings: maxLength and minLength count Unicode code points (supporting emoji and multi-byte characters), and pattern matches ECMA-262 regular expressions through a new RE2 translator. This release includes a reusable translator package for future regex-based keywords like patternProperties.

- internal/ecmaregex/: new package with ECMA-262-to-RE2 translator (translate.go, properties.go), handling shorthands (\d, \w, \s), dot, Unicode escapes, surrogate pairs, named groups, and Unicode properties; rejects lookahead, lookbehind, backreferences, lone surrogates, and unknown properties as schema errors
- internal/keywords/validation/maxlength.go, minlength.go, pattern.go: three self-registering keywords; maxLength/minLength count code points with utf8.RuneCountInString, apply only to strings, and accept 2.0 as integer using exact number model; pattern compiles at compile time and matches unanchored
- testdata/validation/: spec-derived test cases for maxLength (6.3.1), minLength (6.3.2), and pattern (6.3.3), covering ASCII, multi-byte, emoji, non-string instances, translated escapes, surrogate pairs, and Unicode properties
- docs/regex-deviations.md: comprehensive guide to ECMA-262 vs RE2 differences, translated constructs, accepted literals, schema error conditions, supported Unicode properties, and pattern-size limit
- cmd/jsonschema/pattern_test.go, patternsize_test.go: CLI contract tests verifying schema errors exit with code 2 and error messages name the keyword and problem without instance values
- patternsize_test.go: library test verifying pattern-size limit constant (64 KiB default, read from code) is enforced before translation

## How to verify

- maxLength: schema {"maxLength":2} rejects 3-character strings and emoji counts as one character: `jsonschema validate <(echo '{"maxLength":2}') <(echo '"😀😀😀"')` → invalid
- minLength: schema {"minLength":2} accepts emoji pair and rejects single emoji: `jsonschema validate <(echo '{"minLength":2}') <(echo '"😀😀"')` → valid
- pattern with ECMA escapes: schema {"pattern":"^\\d+$"} accepts ASCII digits only, rejects Arabic-Indic, and ^.$ matches single emoji: `jsonschema validate <(echo '{"pattern":"^.$"}') <(echo '"😀"')` → valid
- pattern schema errors: lookahead (?=a), lookbehind (?<=a), backreferences \1 or \k<name>, lone surrogates \ud800, unknown properties \p{scx=Greek} all exit with code 2 and stderr naming 'pattern' and the problem (no instance value)
- pattern-size limit: schema with pattern 64 KiB compiles; 64 KiB + 1 byte is schema error with code 2
- concurrent: no race detector warnings under `go test -race ./...` on pattern evaluation
- non-string instances: maxLength, minLength, pattern all pass (return true) for numbers, objects, arrays, null

## Migration

none

## Rollback

Remove internal/ecmaregex/, internal/keywords/validation/maxlength.go, maxlength_test.go, minlength.go, minlength_test.go, pattern.go, pattern_test.go, testdata/validation/maxLength.json, minLength.json, pattern.json, cmd/jsonschema/pattern_test.go, patternsize_test.go, patternsize_test.go (root), and docs/regex-deviations.md.

## Known limitations

- Pattern-size limit is currently a fixed 64 KiB constant in internal/keywords/validation/pattern.go. Story 193 will add library option and CLI flag to configure it; this change deferred per operator clarification.
- Lookahead (?=), lookbehind (?<=), backreferences (\1, \k<name>), and RE2-only constructs (\z, \A, \Q...\E, (?i) flags, POSIX [[:alpha:]]) are schema errors, never silently translated.
- Unicode properties limited to Go's unicode.Scripts and unicode.Categories tables; Script_Extensions/scx unsupported. Supported names documented in docs/regex-deviations.md.
- Matching is unanchored with Go regexp MatchString; patterns are interpreted per ECMA-262 Unicode mode with no flags.

## Concerns

- Pattern-size option and CLI flag originally planned for this story were moved to story 193 by operator decision (confirmed in approved batch reservation). Constant default is in place and tested; configuration infrastructure deferred.
- Tester identified three LOW findings: possible Unicode property list drift with Go 1.25+ (acceptable, documented set validated); JSON Pointer needed for nested pattern errors in future (hub concern, not this story); Number allocation in length keywords per evaluation (minor efficiency, correctness unaffected). All non-blocking.
