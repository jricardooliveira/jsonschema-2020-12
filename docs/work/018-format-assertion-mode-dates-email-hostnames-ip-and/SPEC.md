# SPEC #18: Format assertion: mode, dates, email, hostnames, IP and UUID

- Date: 2026-10-17
- Type: feature
- Verdict: pass
- Status: passed gate 1 with nothing for the operator to decide (run #153, 2026-10-17 02:32 UTC)

## Problem

'format' is only collected as an annotation today. The brief requires an optional assertion mode with three ways to turn it on: a library option, a CLI flag, or a meta-schema that declares the format-assertion vocabulary as required. In that mode, the 2020-12 formats must actually be checked using only the standard library.

## Why this matters

Format assertion is listed in the brief's 'Optional features'. Callers and evaluators who turn it on expect values such as dates, emails, hostnames, IP addresses and uuids to fail when malformed. The default must stay annotation-only, exactly as 2020-12 requires.

## Acceptance criteria

- Without the mode, 'format' never affects validity. It is still collected as an annotation, exactly as before this story.
- A library compile option turns format-assertion mode on for that compiled schema. A string that fails a known format is then invalid. The error reports the instance location and the keyword location, never the value.
- A CLI flag of 'jsonschema validate' turns the mode on. With the flag, a document with a malformed format value exits 1 and a well-formed one exits 0. Without the flag, both exit 0.
- A schema whose meta-schema lists the format-assertion vocabulary as required turns the mode on. The code reads the flag that story 190 recorded and does not parse $vocabulary again. When the vocabulary is listed as optional and the caller did not turn the mode on, format stays annotation-only.
- In the mode, unknown formats and the not-yet-built 2020-12 formats (uri, uri-reference, iri, iri-reference, uri-template, json-pointer, relative-json-pointer, regex) always pass and are never an error.
- In the mode, format checks apply only to string instances. Every other type passes.
- Format checks have one registration point, internal/formats/registry.go. Each check lives in its own file internal/formats/<format>.go and registers itself from init(). No central list or switch is edited to add a format.
- date-time, date and time strictly follow RFC 3339 section 5.6. This covers real calendar days, including leap years, and a required time offset. Second 60 is accepted only when it falls at 23:59:60 UTC after the offset is applied. A lowercase 't' or 'z' is valid.
- duration follows RFC 3339 appendix A. P3D, PT1H30M and P2W pass. P1W3D, P and PT fail.
- email follows the full RFC 5321 address rules. This includes dot-atom and quoted local parts, and domain literals such as user@[192.168.0.1] and user@[IPv6:...]. Each part has a length limit.
- hostname follows RFC 1123. Each label is 1-63 letters, digits or hyphens, with no leading or trailing hyphen. The total length is at most 253. A trailing dot fails.
- idn-hostname and idn-email get basic checks only. These are the overall and per-label length, no leading or trailing hyphen, and no spaces, control characters or stray punctuation. Plain-ASCII input follows the hostname and email rules. Anything that needs the Unicode tables is accepted, and the code records what is not checked.
- ipv4 is a strict dotted-quad (RFC 2673): exactly four decimal parts, each 0-255, with no leading zeros, so 192.168.01.1 fails. ipv6 follows RFC 4291 text forms, including '::' compression and an embedded IPv4 tail. A zone suffix such as fe80::1%eth0 fails.
- uuid checks shape only: 8-4-4-4-12 hexadecimal digits in either case. Any version or variant digit is accepted.
- With the mode on, meta-schema validation of a schema never asserts format. A schema with "$id": "http://example.com/a b" still compiles with the mode on and with the mode off.
- Every format has positive and negative tests, and so does each way of turning the mode on. go test -race ./... passes, and go.mod gains no require lines.
- docs/formats.md lists the formats checked so far, what idn-hostname and idn-email do not check, and that the not-yet-built formats pass until a later story adds them.

## Non-goals

- The uri, uri-reference, iri, iri-reference, uri-template, json-pointer, relative-json-pointer and regex checks. Story 196 adds these.
- Full IDNA/Unicode validation for idn-hostname and idn-email, because golang.org/x/net/idna was declined.
- Asserting format when the format-assertion vocabulary is only optional.
- The complete README/CLI documentation, which story 197 covers.
- Any change to output shapes or exit-code rules.

## Tasks

- **T1** — Format registry and assertion-mode plumbing
  - purpose: Create the single format registration point and turn the mode on from the library option or from story 190's required-vocabulary flag. Make the format keyword assert only in the mode, and keep meta-schema checks annotation-only.
  - scope: internal/formats/registry.go, options.go, internal/compile/compile.go, internal/compile/metacheck.go, internal/keywords/formatannotation/format.go, internal/keywords/formatannotation/assert_test.go
  - done when: assert_test.go shows four things: a registered stub format fails in the mode and passes outside it; unknown formats pass; non-strings pass; and the meta-check compiles a schema with mode on.
  - delivers: AC1, AC2, AC4, AC5, AC6, AC7, AC15
- **T2** — Date and time formats
  - purpose: Add strict RFC 3339 checks for date-time, date, time and duration, each in its own file with its own tests.
  - scope: internal/formats/datetime.go, internal/formats/datetime_test.go, internal/formats/date.go, internal/formats/date_test.go, internal/formats/time.go, internal/formats/time_test.go, internal/formats/duration.go, internal/formats/duration_test.go
  - done when: The table tests show these results: leap years and lowercase t/z pass; a non-leap second 60, P1W3D, P and PT fail.
  - depends on: T1
  - delivers: AC8, AC9, AC16
- **T3** — Hostname and email formats, ASCII and international
  - purpose: Add the RFC 1123 hostname check, the full RFC 5321 email check, and basic idn-hostname and idn-email checks that record their gaps in code.
  - scope: internal/formats/hostname.go, internal/formats/hostname_test.go, internal/formats/idnhostname.go, internal/formats/idnhostname_test.go, internal/formats/email.go, internal/formats/email_test.go, internal/formats/idnemail.go, internal/formats/idnemail_test.go
  - done when: The table tests show these results: a trailing dot fails; quoted local parts and IP domain literals pass; hyphen-edged labels fail; ASCII idn input matches the hostname results.
  - depends on: T1, T4
  - delivers: AC10, AC11, AC12, AC16
- **T4** — IP address and uuid formats
  - purpose: Add the strict dotted-quad ipv4 check, the RFC 4291 ipv6 check without zones, and the shape-only uuid check. Expose the IP helpers so the email domain-literal check can reuse them.
  - scope: internal/formats/ipv4.go, internal/formats/ipv4_test.go, internal/formats/ipv6.go, internal/formats/ipv6_test.go, internal/formats/uuid.go, internal/formats/uuid_test.go
  - done when: The table tests show these results: 192.168.01.1 and fe80::1%eth0 fail; an embedded IPv4 tail passes; an upper-case uuid with version digit 0 or 9 passes.
  - depends on: T1
  - delivers: AC13, AC14, AC16
- **T5** — CLI flag and end-to-end enablement tests
  - purpose: Add the format-assertion flag to 'jsonschema validate'. Test every way of turning the mode on through the library and the CLI, including the meta-check case.
  - scope: cmd/jsonschema/run.go, cmd/jsonschema/formatassert_cli_test.go, formatassertion_test.go
  - done when: The CLI tests show exit 1 with the flag and exit 0 without it. The library tests cover the option, a required vocabulary, an optional vocabulary, not-yet-built formats, and the '$id' 'a b' case with mode on and off. All pass under -race.
  - depends on: T1, T2, T3, T4
  - delivers: AC2, AC3, AC4, AC5, AC15, AC16
- **T6** — Document checked and partly checked formats
  - purpose: Record which formats are checked, what the idn checks miss, and that the remaining formats pass for now.
  - scope: docs/formats.md
  - done when: docs/formats.md lists all 11 checked formats, the idn gaps and the 8 deferred formats.
  - depends on: T3
  - delivers: AC17
