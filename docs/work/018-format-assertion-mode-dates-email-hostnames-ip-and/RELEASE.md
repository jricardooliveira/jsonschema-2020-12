# Release notes — #18: Format assertion: mode, dates, email, hostnames, IP and UUID

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #18
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

Adds optional format assertion mode that validates 11 JSON formats (dates, times, email, hostnames, IP addresses, UUID) against their RFC standards. Default behavior is unchanged (format annotation-only); users opt in via library option, CLI flag, or required meta-schema vocabulary. When enabled, strings failing format checks become invalid; unknown and not-yet-built formats always pass.

- internal/formats/: new package with registry and 11 RFC-compliant checkers (date, date-time, time, duration, email, idn-email, hostname, idn-hostname, ipv4, ipv6, uuid), each in its own file registering from init()
- internal/keywords/formatannotation/format.go: format keyword now asserts validity when mode is on, still collects annotation in both modes
- options.go: new FormatAssertion boolean field (default false)
- internal/compile/refs.go: carries format assertion mode through compilation; reads story 190's recorded vocabulary flag per resource
- internal/compile/metacheck.go: forces format assertion off for meta-schema validation of schemas
- cmd/jsonschema/run.go: new --format-assertion flag on 'jsonschema validate' (default false)
- docs/formats.md: lists 11 checked formats with RFC citations, idn-* basic-only gaps, and 8 deferred formats

## How to verify

- Library: compile with Options{FormatAssertion: true}, validate a document with format:date and value '2023-02-29', validation fails; without the option, it passes
- Library: compile with Options{FormatAssertion: true}, validate format:ipv4 with '192.168.01.1' (leading zero), fails; without option, passes
- CLI: 'jsonschema validate --format-assertion schema.json doc.json' with malformed format value exits 1; same without flag exits 0
- Meta-schema: compile a schema whose meta-schema requires the format-assertion vocabulary; invalid formats in the schema fail validation
- Meta-schema: if vocabulary is optional, format stays annotation-only unless caller set the option
- Unknown formats: with --format-assertion, a document with format:no-such-format always passes
- Non-strings: format validation applies only to strings; integers, arrays etc. pass regardless of mode
- Meta-check: compile a schema with $id 'http://example.com/a b' (violates uri-reference) in both modes; succeeds in both because meta-schema checks force assertion off

## Migration

none

## Rollback

Set Options.FormatAssertion to false (default) or omit --format-assertion to revert all schemas to annotation-only. No data or schema migration needed.

## Known limitations

- idn-hostname and idn-email: basic checks only (UTF-8 validity, length, per-label length, hyphen edges, no spaces/control chars). No IDNA2008 tables, bidi rules, normalization, or RFC 6531 categories (stdlib only, golang.org/x/net/idna declined).
- Not yet built (always pass in assertion mode): uri, uri-reference, iri, iri-reference, uri-template, json-pointer, relative-json-pointer, regex. Story 196 implements these.
- Duration format accepts non-adjacent unit combinations that RFC 3339 appendix A forbids (e.g., P1Y3D year-day without month, PT1H5S hour-second without minute).

## Concerns

- Duration validator violates RFC 3339 appendix A strict grammar: accepts P1Y3D and PT1H5S, which should fail because units must appear in order (Y before M before D, H before M before S). The implementation allows skipping units, passing false positives. This is a medium-severity correctness gap: users expecting RFC-compliant duration validation will get incorrect results. The acceptance criteria explicitly require RFC 3339 compliance.
- Two lower-risk test gaps: (1) no test proves format validation charges bytes to the step budget (ChargeBytes is called but not asserted), and (2) no test covers embedded resources with a different $schema and format-assertion vocabulary setting.
