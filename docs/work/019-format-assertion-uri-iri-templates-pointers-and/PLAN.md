# PLAN #19: Format assertion: URI, IRI, templates, pointers and regex

- Date: 2026-10-17
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-019-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/formats/rfc3986.go`
- `internal/formats/rfc3986_test.go`
- `internal/formats/uri.go`
- `internal/formats/uri_test.go`
- `internal/formats/uri_reference.go`
- `internal/formats/uri_reference_test.go`
- `internal/formats/iri.go`
- `internal/formats/iri_test.go`
- `internal/formats/iri_reference.go`
- `internal/formats/iri_reference_test.go`
- `internal/formats/uri_template.go`
- `internal/formats/uri_template_test.go`
- `internal/formats/json_pointer.go`
- `internal/formats/json_pointer_test.go`
- `internal/formats/relative_json_pointer.go`
- `internal/formats/relative_json_pointer_test.go`
- `internal/formats/ecma_syntax.go`
- `internal/formats/ecma_syntax_test.go`
- `internal/formats/ecma_unicode_props.go`
- `internal/formats/ecma_unicode_props_test.go`
- `internal/formats/regex.go`
- `internal/formats/regex_test.go`

Modules the architecture expects to touch:

- `internal/formats/rfc3986.go`
- `internal/formats/rfc3986_test.go`
- `internal/formats/uri.go`
- `internal/formats/uri_test.go`
- `internal/formats/uri_reference.go`
- `internal/formats/uri_reference_test.go`
- `internal/formats/iri.go`
- `internal/formats/iri_test.go`
- `internal/formats/iri_reference.go`
- `internal/formats/iri_reference_test.go`
- `internal/formats/uri_template.go`
- `internal/formats/uri_template_test.go`
- `internal/formats/json_pointer.go`
- `internal/formats/json_pointer_test.go`
- `internal/formats/relative_json_pointer.go`
- `internal/formats/relative_json_pointer_test.go`
- `internal/formats/ecma_syntax.go`
- `internal/formats/ecma_syntax_test.go`
- `internal/formats/ecma_unicode_props.go`
- `internal/formats/ecma_unicode_props_test.go`
- `internal/formats/regex.go`
- `internal/formats/regex_test.go`
- `internal/formats/assertion_mode_test.go`
- `formatassertion_test.go`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Strict RFC 3986/3987 grammar and the four address formats
2. **T2** — uri-template spelling check
3. **T3** — json-pointer and relative-json-pointer
4. **T5** — Unicode property name list
5. **T4** — Strict ECMA-262 (u-flag) pattern syntax checker  _(after T5)_
6. **T6** — regex format wiring  _(after T4, T5)_

## Risks

- The strict ECMA-262 u-flag grammar is large. Annex B leniencies must stay disabled, and named-group and back-reference rules are subtle. A flat non-recursive scan makes the per-alternative duplicate-name rule depend on careful stack-frame bookkeeping.
- The Unicode property alias list is hand-written. Go's unicode tables may lag behind the Unicode version ECMA-262 references, so some script names may be missing. This is recorded in code.
- Story 182's translation step may reject or mishandle constructs that the syntax checker accepts. It is contained by recover and ignored for the verdict, so its run here only proves the step does not crash.
- formatassertion_test.go (root, outside the reservation) has TestFormatAssertionDeferredAndUnknownPass, which expects these eight formats to pass any string in assertion mode. It breaks once they register. Its list is cut to 'no-such-format' (a one-line edit), and coverage moves to the new internal/formats/assertion_mode_test.go. That new file is also outside the reservation: the in-package test files cannot import the root package because of the import cycle, so an external formats_test file is needed.

Implementation constraints the coder must honor:

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

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: go test -race ./internal/formats/... passes with positive, negative and annotation-only cases for the four formats.
- **T2**: Tests cover matching braces, operators, modifier ranges, varname characters, pct-encoding and annotation-only mode.
- **T3**: Tests cover escapes, leading zeros, '+N'/'-N', '#', and annotation-only mode.
- **T5**: Tests accept 'Letter', 'L', 'Lu', 'Script=Greek', 'sc=Grek', 'Script_Extensions=Latin', 'ASCII' and 'Any', and reject misspellings and unknown script values.
- **T4**: Table tests accept lookaround, back-references and named groups, and reject '\a', a lone '{', unbalanced parentheses, reversed ranges, out-of-range back-references and duplicate group names.
- **T6**: Tests cover valid, invalid and JS-only patterns in assertion mode, and annotation-only behaviour outside it.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
