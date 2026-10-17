# Release notes — #19: Format assertion: URI, IRI, templates, pointers and regex

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #19
- Status: proposed (pending Checkpoint 3 sign-off)

## What changed

In format-assertion mode, the validator now checks eight additional formats (uri, uri-reference, iri, iri-reference, uri-template, json-pointer, relative-json-pointer and regex) that 2020-12 defines. Each format uses a strict grammar from the standard library only, rejects hostile input in linear time, and records any format that is only partly checked. Outside assertion mode, all eight formats remain annotations only.

- internal/formats/rfc3986.go: shared strict RFC 3986/3987 URI and IRI grammar, single-pass iterative scanner
- internal/formats/uri.go, uri_reference.go, iri.go, iri_reference.go: four address format checkers, thin wrappers on rfc3986
- internal/formats/uri_template.go: RFC 6570 URI template spelling check (braces, operators +#./;?&, modifiers :N and *)
- internal/formats/json_pointer.go, relative_json_pointer.go: RFC 6901 and draft-bhutton pointer checkers, with index manipulation (+N/-N)
- internal/formats/ecma_syntax.go: strict ECMA-262 u-flag pattern syntax, non-recursive with explicit group stack, JavaScript-only constructs (lookahead, backreferences) valid
- internal/formats/ecma_unicode_props.go: \p{...} property names built from unicode.Categories and unicode.Scripts plus hand-written aliases
- internal/formats/regex.go: format wiring; verdict from syntax check; advisory Translate call inside recover()
- internal/formats/assertion_mode_test.go (new external package): annotation-only and step-charging tests for all eight formats
- formatassertion_test.go: line 64 edited to remove the eight formats (coverage moved to assertion_mode_test.go)
- Test files added: rfc3986_test.go, uri_test.go, uri_reference_test.go, iri_test.go, iri_reference_test.go, uri_template_test.go, json_pointer_test.go, relative_json_pointer_test.go, ecma_syntax_test.go (includes hostile nesting), ecma_unicode_props_test.go, regex_test.go

## How to verify

- Compile a schema with FormatAssertion on and format: 'uri'; validate 'http://a/b' (pass), '/a' (fail), 'http://a b' (fail with space)
- Compile with format: 'regex'; validate '^a+$' (pass), '\\a' (fail), '(?=a)' (pass, JavaScript lookahead)
- Compile with format: 'json-pointer'; validate '/a~1b' (pass, ~1 is /), '/~' (fail, incomplete escape)
- Compile with format: 'uri-template'; validate '/users/{id}{?page}' (pass), '{' (fail, unmatched)
- Compile with FormatAssertion off (same schema); all invalid values pass (annotation only)
- With MaxSteps 100 and a 64kb string, assertion mode gives LimitSteps error; without assertion mode, it passes
- go test -race ./internal/formats/... passes all table tests, annotation-only tests and hostile input tests (million unclosed parentheses)

## Migration

None. The eight formats were never checked before. In assertion mode they now validate; outside assertion mode (the default) they remain annotations, so all existing schemas work unchanged. Code that uses FormatAssertion: false continues to work. No schema version change, no data migration.

## Rollback

Remove the 23 new files in internal/formats/ (all *_test.go and the checker files), restore formatassertion_test.go line 64 to list the eight formats. The format registration point and step charging remain; they are unaffected.

## Known limitations

- uri/iri/uri-reference/iri-reference: host is checked for reg-name or IP-literal syntax only (no RFC 5952 IPv4 dotted-quad rule, no IDN or bidi rules). Port digits are not range-limited.
- uri-template: only RFC 6570 level 4 grammar is checked. Reserved-for-future operators (= , ! @ |) are rejected. Syntax only, never expanded.
- json-pointer: fully checked per RFC 6901.
- relative-json-pointer: fully checked per draft-bhutton-relative-json-pointer-00.
- regex: lookahead, lookbehind, back-references and named groups are judged by ECMA-262 syntax only (RE2 cannot run them). \p{...} property names are checked against a list built from Go's unicode.Categories and unicode.Scripts; script names newer than Go's Unicode version or missing aliases fail closed.
- All eight formats have code comments (ponytail) listing what they do not fully check.

## Concerns

- formatassertion_test.go line 64: one-line edit removes the eight formats from a test that expected them to reject all strings in assertion mode (the old behaviour). Now that they are registered and validate strictly, that test would break. Coverage moved to assertion_mode_test.go. This is correct and documented in the architecture notes.
- assertion_mode_test.go (new external package formats_test): necessary because in-package test files cannot import the root package (import cycle). This is the standard Go solution and does not affect the tested code.
- ECMA-262 u-flag grammar is large and complex (lookahead, backreferences, named groups, class ranges, surrogates). The tester confirmed the scanner handles surrogate-pair escapes correctly, rejects hostile nesting (million unclosed parentheses) without crash, and enforces duplicate group names. The implementation uses an explicit heap stack frame per group and a flat class state, no recursion on input length.
- Unicode property aliases are hand-written (scriptAliases in ecma_unicode_props.go). Gaps are recorded in code. This is unavoidable with standard-library-only and is documented in the acceptance criteria.
- The translation step (story 182) is advisory: ecmaregex.Translate is called only after the syntax check accepts the pattern and inside recover(). Its result is ignored. A panic or error becomes 'not translatable' and does not affect the verdict. The regex format stays valid.
