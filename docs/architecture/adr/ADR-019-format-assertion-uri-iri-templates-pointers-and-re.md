# ADR #19: Format assertion: URI, IRI, templates, pointers and regex

- Date: 2026-10-17
- Story: #19
- Status: proposed (pending architecture sign-off)

## Decision
This is the reviewed design with the reviewer's four safe choices applied. Nothing else changes. Before writing it I checked these facts against the code: internal/formats/registry.go has `func Register(name string, check func(string) bool)`; ecmaregex.Translate(pattern string) (string, error) and ecmaregex.DefaultMaxPatternSize exist in internal/ecmaregex/translate.go; keywords.BytesPerStep = 64 and EvalContext.ChargeBytes are in internal/keywords/registry.go; internal/keywords/formatannotation/format.go calls ctx.ChargeBytes(len(s)) right before c.check(s), and only when a check is set (assertion mode on and a string instance); formatassertion_test.go line 64 lists the eight formats in TestFormatAssertionDeferredAndUnknownPass.

All work goes in new files in internal/formats. Each format registers from init() through formats.Register(name, func(string) bool). registry.go is only read. rfc3986.go is the strict grammar that uri, uri-reference, iri and iri-reference share, with an IRI switch. The regex format is judged by ecma_syntax.go together with ecma_unicode_props.go.

Reviewer points:
(1) The ECMA-262 u-flag syntax checker does not recurse. It makes one left-to-right scan over the pattern's runes. It keeps an explicit heap slice as a group stack: one frame per open '(' holds the group kind and the named groups of the current alternative. It also keeps a flat 'inside a class' state, because classes do not nest under the u flag. The same pass collects group names, the group count and named back-references. Forward references are checked once after the scan, against the collected sets. Each rune is visited a constant number of times, so the check is linear in input length and needs no depth cap. ecma_syntax_test.go feeds one million '(' and one million '[' characters. Each must return invalid quickly, with no stack growth and no crash.
(2) rfc3986.go, uri_template.go, json_pointer.go and relative_json_pointer.go are single-pass scanners that walk the input by index. They never recurse over input length and never run a regexp over the input. The RFC 3986 hier-part, authority, IP-literal, path, query and fragment rules are helper calls of fixed depth. Each one moves forward through the input and looks ahead at most a few characters, such as '//' or the ':' after a scheme. This also replaces the 'recursive-descent' wording in story task T1: the grammar is the same, the control flow is iterative.
(3) In regex.go the translation step is advisory only, and the verdict comes from ecma_syntax.go alone. regex.go calls ecmaregex.Translate through one small pure helper, translateContained(tr func(string) (string, error), s string) bool:
- it uses defer/recover, so any panic becomes 'not translatable';
- it ignores errors;
- it runs only when the syntax check has already accepted the pattern and len(s) <= ecmaregex.DefaultMaxPatternSize, which bounds the expanded RE2 text;
- it never calls regexp.Compile, so no RE2 compile memory is spent here;
- its result is thrown away. A ponytail comment says the call exists only to exercise the reused step, so it can never turn a valid pattern into a failure.
regex_test.go passes a panicking function and an error-returning function to translateContained. It shows that neither crashes and that the regex check still accepts '(?=a)', '(a)\\1' and '(?<n>a)'.
(4) Step charging is already central: format.go charges ctx.ChargeBytes(len(s)) before every check in assertion mode (ADR-016 H, story 193). Each new checker makes a constant number of linear passes; regex makes at most the syntax pass plus Translate. That one charge therefore stays in proportion to the work, and no checker file needs its own charging. A new test proves this for each of the eight formats: a string of 1000*keywords.BytesPerStep bytes with MaxSteps 100 gives a LimitSteps error in assertion mode and does not outside it.

Annotation-only tests: the existing in-package tests (package formats, helper table() in datetime_test.go) cannot import the root package because of the import cycle. The annotation-only and step-charging tests for all eight formats therefore go in one new external-package file, internal/formats/assertion_mode_test.go (package formats_test). It is table-driven, with one invalid sample per format, and compiles each schema with FormatAssertion on and off.

Existing test that breaks: in formatassertion_test.go, TestFormatAssertionDeferredAndUnknownPass expects these eight formats to accept any string in assertion mode. Its list (line 64) shrinks to 'no-such-format' only. This is a one-line edit; the coverage moves to assertion_mode_test.go.

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
- The ECMA-262 syntax checker is non-recursive: one forward scan, with an explicit heap slice as the group stack and a flat class state. It is linear in pattern length and has no depth cap. Named and numbered back-references are resolved after the scan from the collected sets.
- The RFC 3986/3987, RFC 6570, json-pointer and relative-json-pointer checkers are single-pass index scanners. They never recurse over input length and never run a regexp over the input.
- regex.go calls ecmaregex.Translate only through translateContained (defer/recover; errors and panics are ignored), and only after the syntax check has accepted the pattern and len(s) <= ecmaregex.DefaultMaxPatternSize. It never calls regexp.Compile. The verdict comes only from the syntax checker.
- Step charging stays in formatannotation/format.go (ctx.ChargeBytes(len(s))). Checkers must make a constant number of linear passes, so that charge stays proportional, and must not keep their own counters.
- Check functions are pure and keep no shared mutable state.
- Each partly checked format has a comment that says what it misses.
- Tests ask the code for every defined limit instead of repeating the number: the uri-template max prefix 9999 as a named constant, keywords.BytesPerStep and ecmaregex.DefaultMaxPatternSize.
- Format errors never quote the instance value.
- ecma_syntax_test.go tests one million '(' and one million '[' characters: both are invalid, fast, with no crash. regex_test.go tests translateContained with a panicking translator and with an error-returning translator, and shows the verdict does not change.

## Risks
- The strict ECMA-262 u-flag grammar is large. Annex B leniencies must stay disabled, and the named-group and back-reference rules are subtle. Because the scan is flat and non-recursive, the rule against duplicate names in one alternative depends on careful bookkeeping in the stack frames.
- The Unicode property alias list is hand-written. Go's unicode tables may lag behind the Unicode version ECMA-262 references, so some script names may be missing. The gap is recorded in code.
- Story 182's translation step may reject or mishandle constructs that the syntax checker accepts. Its run here is contained by recover and ignored for the verdict, so it only proves that the step does not crash.
- formatassertion_test.go is in the root package, outside the reservation. Its TestFormatAssertionDeferredAndUnknownPass (line 64) expects these eight formats to accept any string in assertion mode, so it breaks once they register. Its list is cut to 'no-such-format' (a one-line edit) and the coverage moves to the new internal/formats/assertion_mode_test.go.
- internal/formats/assertion_mode_test.go is also outside the reservation. It is needed because the in-package test files cannot import the root package (import cycle), so the annotation-only tests need an external formats_test file.

## Breaking changes
- none

## Sensitivity
- security
