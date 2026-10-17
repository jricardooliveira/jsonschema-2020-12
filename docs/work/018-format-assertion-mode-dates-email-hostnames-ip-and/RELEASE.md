# Release notes — #18: Format assertion: mode, dates, email, hostnames, IP and UUID

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #18
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

Format assertion is now an optional mode: the 'format' keyword fails validation for malformed values only when explicitly enabled via a library option, a CLI flag, or a meta-schema declaring the format-assertion vocabulary as required. 11 formats are now checked (date-time, date, time, duration, email, idn-email, hostname, idn-hostname, ipv4, ipv6, uuid) using only the standard library, strictly following RFC 3339, RFC 5321, RFC 1123, RFC 2673, RFC 4291 and RFC 4122. The default stays annotation-only; unknown formats are never an error.

- internal/formats/: new package with registry (registry.go) and 11 format checkers in their own files (date.go, datetime.go, time.go, duration.go, email.go, idnemail.go, hostname.go, idnhostname.go, ipv4.go, ipv6.go, uuid.go), each registering from init()
- options.go: new FormatAssertion bool field to enable assertion mode for a compiled schema
- internal/keywords/formatannotation/format.go: rewritten to evaluate format in assertion mode (fail on check failure for strings) while keeping annotation in both modes
- internal/keywords/registry.go: added FormatAssert bool field to CompileContext
- internal/compile/refs.go: threaded FormatAssert boolean through compile context; reads story 190's required-vocabulary flag for per-resource enablement; new CompileReachedFormat function
- internal/compile/metacheck.go: meta-schema validation now explicitly calls compileDocumentOpts with noFormat=true to force annotation-only
- cmd/jsonschema/run.go: new --format-assertion boolean flag for 'validate' subcommand; passes option to compile
- jsonschema.go: calls CompileReachedFormat instead of CompileReachedWith, passing opts.FormatAssertion
- docs/formats.md: new; lists 11 checked formats with RFC citations, documents idn gaps, notes 8 deferred formats
- testdata/vocabulary/required_optional_unknown.json: one test case updated (format-assertion required now asserts format, so 'not-an-email' fails where it previously passed annotation-only)

## How to verify

- Compile with Options{FormatAssertion: true}; '{"format":"ipv4"}' rejects instance '"192.168.01.1"' (invalid: leading zero) but accepts '"192.168.1.1"'
- Without FormatAssertion, both instances return valid
- CLI: 'jsonschema validate --format-assertion schema.json' with an email format and a malformed email instance exits 1; without the flag, exits 0
- A schema whose meta-schema lists the format-assertion vocabulary as required (true in $vocabulary) now asserts format in mode; optional (false) stays annotation-only, so unknown formats and all other defaults apply unless FormatAssertion option is also set
- Unknown formats (uri, regex, json-pointer, etc.) pass in assertion mode; the docs list which are deferred
- 'go test -race ./...' passes; go.mod has no new dependencies

## Migration

None required. Format assertion is opt-in and disabled by default. Existing code sees no change unless Options{FormatAssertion: true} is set or --format-assertion is passed. Format annotations continue to be collected regardless of mode. Schemas with format values are evaluated the same way unless the caller explicitly enables assertion.

## Rollback

Remove Options{FormatAssertion: true} from library calls or omit --format-assertion from CLI invocations. Format keyword returns to annotation-only. No data migration; the compiled schema is immutable and the mode is fixed at compile time.

## Known limitations

- idn-hostname and idn-email are basic checks only: total length, per-label length, hyphen rules, no spaces/control characters, ASCII characters limited to letters/digits/hyphen. They do not validate IDNA2008 categories, normalization or bidi rules. See docs/formats.md for full gaps.
- The formats uri, uri-reference, iri, iri-reference, uri-template, json-pointer, relative-json-pointer and regex are not yet checked; they pass in assertion mode as unknown formats until story 196 implements them.
- Format validation charges against the per-call MaxSteps budget (each string costs steps proportional to its length).
- Error reports for format failure give the instance location and keyword location only; the value is never reported to avoid logging sensitive data.

## Concerns

- F1 (medium): The guard forcing meta-schema validation to annotation-only (noFormat=true in metacheck.go) is exercised by formatassertion_test.go::TestFormatAssertionMetaCheckAnnotationOnly, but the test only verifies compile succeeds. It does not fail without the guard; it would pass if the guard were removed and the meta-schema itself did not assert format-assertion on keywords like $id. An explicit test verifying the guard's effect (e.g., mocking or inspecting the compile path) would strengthen assurance that meta-checks stay annotation-only in all modes.
- F3 (low): No test covers an embedded resource with a different $schema than the root, which could have a different format-assertion vocabulary flag recorded by story 190. This edge case is not in the acceptance criteria but is a potential interaction point.
- F2 (low): IPv6 address literals in email must use the 'IPv6:' prefix per RFC 5321 section 4.1.3; user@[::1] is invalid, user@[IPv6:::1] is valid. This is standard-conformant but differs from some real-world email parsers that tolerate the bare form.
