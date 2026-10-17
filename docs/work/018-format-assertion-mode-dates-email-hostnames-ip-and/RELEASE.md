# Release notes — #18: Format assertion: mode, dates, email, hostnames, IP and UUID

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #18
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

Format assertion is now optional and defaults off. Enable it via library option Options.FormatAssertion=true, CLI flag --format-assertion, or by a schema's meta-schema declaring the format-assertion vocabulary as required. In assertion mode, strings that fail one of 11 new format checks (date-time, date, time, duration, email, idn-email, hostname, idn-hostname, ipv4, ipv6, uuid) are invalid; all other instance types pass. Format is collected as an annotation in both modes, with the mode fixed at compile time.

- internal/formats/: new package with 11 format checkers (each format in its own file: datetime.go, date.go, time.go, duration.go, email.go, idn-email.go, hostname.go, idn-hostname.go, ipv4.go, ipv6.go, uuid.go), each test in its own _test.go file, plus registry.go as the single registration point
- internal/keywords/formatannotation/format.go: rewritten to support assertion mode; now compiles to capture the mode flag and looked-up format check; still collects annotation in both modes
- internal/compile/refs.go: added FormatAssert boolean to compilation context; added compileDocumentOpts to carry both the caller's format option and a noFormat flag for meta-schema checks
- internal/compile/metacheck.go: meta-schema validation is forced to annotation-only by passing noFormat=true so that invalid $id values (like uri-reference violations) never cause compile failure
- internal/keywords/registry.go: added FormatAssert boolean to CompileContext
- numberlimits.go: added Options.FormatAssertion boolean field, defaults to false
- cmd/jsonschema/run.go: added --format-assertion CLI flag (boolean, defaults false); wires to opts.FormatAssertion
- docs/formats.md: new documentation listing 11 checked formats with RFC citations, gaps in idn-hostname and idn-email checks, and 8 deferred formats (uri, uri-reference, etc.) that pass for now
- testdata/vocabulary/required_optional_unknown.json: updated test case 'format-assertion required is recorded' to expect format assertion to now trigger; 'not-an-email' now correctly fails, and 'a@example.com' passes

## How to verify

- Library: compile {"format":"ipv4"} with FormatAssertion true; validate the string '"192.168.01.1"' — must be invalid (RFC 2673 forbids leading zeros)
- Library: same compile; validate the integer 5 against the same schema — must be valid (non-string types always pass even in assertion mode)
- Library: compile {"format":"date"} with FormatAssertion false; validate '"not-a-date"' — must be valid (annotation-only mode)
- CLI: jsonschema validate --format-assertion -s '{"format":"email"}' << 'not-an-email' — exit code must be 1
- CLI: same without --format-assertion — exit code must be 0
- CLI: jsonschema validate --format-assertion -s '{"\$id":"http://example.com/a b"}' << '{}' — exit code 0 or 1 (never 2); the malformed $id is not asserted because meta-schema checks stay annotation-only
- Run go test -race ./... — all tests pass, no race conditions on the compiled schema or evaluation context

## Migration

None. Format assertion is new and opt-in with default-off. Existing code needs no changes; format remains an annotation until explicitly enabled.

## Rollback

Remove FormatAssertion from Options (numberlimits.go), remove the --format-assertion flag registration (cmd/jsonschema/run.go), and revert internal/keywords/formatannotation/format.go to annotation-only. The internal/formats/ package can remain or be deleted; it is inert if the format keyword does not call Lookup.

## Known limitations

- Eight 2020-12 format checks are deferred to story 196: uri, uri-reference, iri, iri-reference, uri-template, json-pointer, relative-json-pointer, regex. These pass in assertion mode and are listed in docs/formats.md.
- idn-hostname and idn-email use basic checks only (no Unicode tables from stdlib): per-label and total length, no leading/trailing hyphens, no spaces/control characters, ASCII limited to letters/digits/hyphens. Full IDNA2008 validation (code point categories, normalization, bidi) is not performed and documented in the code comments.
- All 8 deferred formats and all unknown format names always pass in assertion mode and are never an error.

## Concerns

- Email IPv6 literal tag is case-sensitive (medium severity): only uppercase 'IPv6:' in addresses like user@[IPv6:...] is accepted; lowercase 'ipv6:' is rejected. RFC 5321 section 4.1.3 specifies 'IPv6:' in that capitalization, but some mailers may emit lowercase variants, causing valid addresses to fail here.
- idn-email rejects quoted local parts when the domain contains non-ASCII (low severity): email '"john doe"@example.com' is valid (delegated to validEmail), but '"john doe"@例え.jp' is invalid because the non-ASCII path applies space/quote character checks to the quoted string. The quoted local part is valid per RFC 5321 even with a non-ASCII domain.
- Email 254-byte total limit has no standalone boundary test (low severity): the limit is implemented but coverage is only indirect through other cases. No test verifies the exact boundary (253 bytes valid, 255 invalid, or vice versa).
