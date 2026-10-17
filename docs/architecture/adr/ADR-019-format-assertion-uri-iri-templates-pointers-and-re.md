# ADR #19: Format assertion: URI, IRI, templates, pointers and regex

- Date: 2026-10-17
- Story: #19
- Status: proposed (pending architecture sign-off)

## Decision
This is the previous design with the reviewer's four points applied. All work still goes in new files in internal/formats (story 195). Each format registers from init() through formats.Register(name, func(string) bool) in internal/formats/registry.go, which is read and never edited. rfc3986.go is the strict grammar shared by uri, uri-reference, iri and iri-reference, with an IRI switch. The regex format is judged by ecma_syntax.go together with ecma_unicode_props.go.

Reviewer points:
(1) The ECMA-262 u-flag syntax checker does not recurse. It is one left-to-right scan over the pattern's runes. It keeps an explicit heap slice as a group stack: one frame per open '(' holds the group kind and a per-alternative set of named groups. It also keeps a flat 'inside a class' state, because classes do not nest under the u flag. Group names, the group count and named back-references are collected in the same pass. Forward references are checked once, after the scan, against the collected sets. Each rune is visited a constant number of times, so the check is linear in input length and needs no depth cap. ecma_syntax_test.go feeds one million '(' and one million '[' characters. Each must return invalid quickly, with no stack growth and no crash.
(2) rfc3986.go, uri_template.go, json_pointer.go and relative_json_pointer.go are single-pass, index-based scanners. They have no recursion over input length and use no regexp over the input. The RFC 3986 hier-part, authority, IP-literal, path, query and fragment rules are fixed-depth helper calls. Each consumes input forward and never backtracks more than a bounded lookahead, such as '//' or the scheme ':' probe.
(3) In regex.go the translation step is advisory only. The verdict is ecma_syntax.go's alone. regex.go calls ecmaregex.Translate through a small pure helper, translateContained(tr func(string) (string, error), s string) bool. The helper uses defer/recover, so any panic becomes 'not translatable', and it ignores errors. It calls Translate only when the syntax check already accepted the pattern and len(s) <= ecmaregex.DefaultMaxPatternSize, which bounds the expanded RE2 text. It never calls regexp.Compile, so RE2 compile memory is never spent here. The result is thrown away apart from a ponytail comment saying it exists only to exercise the reused step; it can never flip a valid pattern to invalid. regex_test.go passes a panicking function and an error-returning function to translateContained. It shows neither crashes, and that the regex check still accepts '(?=a)', '(a)\\1' and '(?<n>a)'.
(4) Step charging is already central. internal/keywords/formatannotation/format.go calls ctx.ChargeBytes(len(s)) before every check in assertion mode (ADR-016 H, story 193). Each new checker does a constant number of linear passes; regex does at most syntax plus Translate. So that one charge stays proportional to the work, and no checker file needs its own charging. A new test proves this for each of the eight formats: a string of 1000*keywords.BytesPerStep bytes with MaxSteps 100 is a LimitSteps error in assertion mode and is not one outside it.

Annotation-only tests: the existing in-package tests (package formats, helper table() in datetime_test.go) cannot import the root package because of the import cycle. So the annotation-only and step-charging tests for all eight formats go in one new external-package file, internal/formats/assertion_mode_test.go (package formats_test). It is table-driven, with one invalid sample per format, and compiles with FormatAssertion on and off.

Existing test that breaks: formatassertion_test.go TestFormatAssertionDeferredAndUnknownPass asserts that these eight formats pass '%%% not anything' in assertion mode. That list shrinks to 'no-such-format' only. This is a one-line edit; the coverage moves to assertion_mode_test.go.

No crowded file is touched.

## Affected modules
- internal/formats/rfc3986.go
- internal/formats/rfc3986_test.go
- internal/formats/uri.go
- internal/formats/uri_test.go
- internal/formats/uri_reference.go
- internal/formats/uri_reference_test.go
- internal/formats/iri.go
- internal/formats/iri_test.go
- internal/formats/iri_reference.go
- internal/formats/iri_reference_test.go
- internal/formats/uri_template.go
- internal/formats/uri_template_test.go
- internal/formats/json_pointer.go
- internal/formats/json_pointer_test.go
- internal/formats/relative_json_pointer.go
- internal/formats/relative_json_pointer_test.go
- internal/formats/ecma_syntax.go
- internal/formats/ecma_syntax_test.go
- internal/formats/ecma_unicode_props.go
- internal/formats/ecma_unicode_props_test.go
- internal/formats/regex.go
- internal/formats/regex_test.go
- internal/formats/assertion_mode_test.go
- formatassertion_test.go

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- Standard library only; go.mod stays without require lines.
- Do not use net/url to judge validity: it is lenient (it accepts 'http://exa mple.com').
- Never edit internal/formats/registry.go, internal/keywords/registry.go, internal/keywords/formatannotation/format.go, internal/ecmaregex/translate.go, internal/eval/eval.go, jsonschema.go, options.go or cmd/jsonschema/*.
- The ECMA-262 syntax checker is non-recursive: one forward scan with an explicit heap slice as the group stack and flat class state. It is linear in pattern length and has no depth cap. Named and numbered back-references are resolved after the scan from collected sets.
- The RFC 3986/3987, RFC 6570, json-pointer and relative-json-pointer checkers are single-pass index scanners with no recursion over input length and no regexp over the input.
- regex.go calls ecmaregex.Translate only through translateContained (defer/recover; errors and panics are ignored), and only after the syntax check accepted the pattern and len(s) <= ecmaregex.DefaultMaxPatternSize. It never calls regexp.Compile. The verdict comes only from the syntax checker.
- Step charging stays in formatannotation/format.go (ctx.ChargeBytes(len(s))). Checkers must do a constant number of linear passes, so that charge stays proportional, and must not keep their own counters.
- Check functions are pure and keep no shared mutable state.
- Each partly checked format has a comment that says what it misses.
- Tests ask the code for any defined limit (the uri-template max prefix 9999 as a named constant, keywords.BytesPerStep, ecmaregex.DefaultMaxPatternSize) instead of repeating the number.
- Format errors never quote the instance value.
- ecma_syntax_test.go tests one million '(' and one million '[' characters: both invalid, fast, no crash. regex_test.go tests translateContained with a panicking and an error-returning translator and shows the verdict is unchanged.

## Risks
- The strict ECMA-262 u-flag grammar is large. Annex B leniencies must stay disabled, and named-group and back-reference rules are subtle. A flat non-recursive scan makes the per-alternative duplicate-name rule depend on careful stack-frame bookkeeping.
- The Unicode property alias list is hand-written. Go's unicode tables may lag behind the Unicode version ECMA-262 references, so some script names may be missing. This is recorded in code.
- Story 182's translation step may reject or mishandle constructs that the syntax checker accepts. It is contained by recover and ignored for the verdict, so its run here only proves the step does not crash.
- formatassertion_test.go (root, outside the reservation) has TestFormatAssertionDeferredAndUnknownPass, which expects these eight formats to pass any string in assertion mode. It breaks once they register. Its list is cut to 'no-such-format' (a one-line edit), and coverage moves to the new internal/formats/assertion_mode_test.go. That new file is also outside the reservation: the in-package test files cannot import the root package because of the import cycle, so an external formats_test file is needed.

## Breaking changes
- none

## Sensitivity
- security
