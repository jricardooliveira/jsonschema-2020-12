# Release notes — #18: Format assertion: mode, dates, email, hostnames, IP and UUID

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #18
- Status: proposed (pending Checkpoint 3 sign-off)

## What changed

Adds optional format assertion mode with 11 RFC-compliant format checkers (date-time, date, time, duration, email, hostname, ipv4, ipv6, idn-email, idn-hostname, uuid). The mode is off by default; formats remain annotations only unless the caller enables assertion via library option, CLI flag, or meta-schema. Default behaviour and invalid-schema handling are unchanged.

- internal/formats/: new registry and 11 format checkers, each in its own file with tests, RFC-strict with boundary tests
- internal/keywords/formatannotation/format.go: extended to compile and evaluate format assertion; annotates in both modes
- options.go: added Options.FormatAssertion boolean field
- internal/compile/refs.go: added FormatAssert and formatOpt fields to shared; compileDocumentOpts passes them through
- internal/compile/metacheck.go: forces format assertion off (noFormat=true) for meta-schema checks
- internal/keywords/registry.go: added FormatAssert field to CompileContext
- cmd/jsonschema/run.go: added --format-assertion CLI flag, wired to opts.FormatAssertion
- jsonschema.go: calls CompileReachedFormat with opts.FormatAssertion
- docs/formats.md: new, lists 11 checked formats, idn gaps, and 8 deferred formats
- testdata/vocabulary/required_optional_unknown.json: updated test data (format-assertion required now rejects malformed email; optional does not)
- Tests: assert_test.go, formatassert_cli_test.go, formatassertion_test.go cover enablement paths and assertion behaviour

## How to verify

- Compile {"format":"email"} with Options.FormatAssertion=true, validate "not@valid" → invalid; without the option → valid
- CLI: jsonschema validate --format-assertion schema.json doc.json → exit 1 if doc has malformed format; exit 0 without --flag
- Meta-schema: a schema with $schema pointing to a meta that lists format-assertion as required → assertion is on; optional listing → off
- Non-strings: validate 5 against {"format":"email"} with assertion on → valid (only strings are checked)
- Deferred formats: uri, regex in assertion mode → pass (not yet built)
- Meta-check: compile {"$id":"http://example.com/a b"} with assertion on and off → both succeed (meta-schema never asserts format)

## Migration

none

## Rollback

The feature is disabled by default and additive. To rollback, do not use Options.FormatAssertion=true or the --format-assertion flag. No schema or instance changes required.

## Known limitations

- idn-hostname and idn-email get basic checks only: length, no leading/trailing hyphens, no spaces/control chars, ASCII limits. Full Unicode/IDNA tables not checked (stdlib only).
- These formats not yet built and always pass in assertion mode: uri, uri-reference, iri, iri-reference, uri-template, json-pointer, relative-json-pointer, regex.
- Meta-schema validation of a schema never asserts format, even if the caller enabled the mode, so a schema with $id containing spaces compiles in all modes.

## Concerns

- Leap-second rule (23:59:60 UTC only) applies offset-adjusted UTC; tested for positive and negative offsets.
- RFC 5321 email: quoted-string and domain-literal parsing is intricate; length limits (64-byte local part, 254 bytes overall) applied after parsing; fully tested with boundary cases.
- Format assertion mode is decided at compile time and read from story 190's per-resource recorded vocabulary flag. If the flag is per-meta-schema rather than per-resource, embedded resources with a different $schema must be tested separately; tests confirm per-resource behaviour.
- Depends on story 190's vocabulary flag being correctly recorded in vocab.FormatAssertion and read without re-parsing $vocabulary.
