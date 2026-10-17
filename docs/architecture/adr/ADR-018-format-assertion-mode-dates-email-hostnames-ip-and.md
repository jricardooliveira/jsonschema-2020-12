# ADR #18: Format assertion: mode, dates, email, hostnames, IP and UUID

- Date: 2026-10-17
- Story: #18
- Status: passed gate 2 with nothing for the operator to decide (run #153, 2026-10-17 02:32 UTC)

## Decision
Package internal/formats holds the registry: Register(name string, check func(string) bool) and Lookup(name) (check, ok). Each format file registers from init(). Because formatannotation/format.go imports internal/formats, every check is linked in without editing internal/keywords/all/all.go. The mode is decided once at compile time. Either the library option (options.go, carried into internal/compile/compile.go's compile options) or story 190's recorded required-format-assertion vocabulary flag sets an effective FormatAssert boolean for that schema resource. format.go's compile function reads that boolean and captures the looked-up check in its compiled, immutable keyword value. At evaluation time, a non-string instance or a missing check passes. A failing check produces an error unit without the value. The annotation is still collected in both modes. The meta-schema check in internal/compile/metacheck.go always compiles and evaluates meta-schemas with FormatAssert forced off, so the 'uri-reference' format on $id is never asserted. Email domain literals reuse the ipv4/ipv6 helpers. idn-hostname reuses the hostname label rules for ASCII input. Format tests are Go table tests in each format's own _test.go file, with citations to the RFC or spec section.

## Affected modules
- internal/formats/registry.go
- internal/formats/datetime.go
- internal/formats/datetime_test.go
- internal/formats/date.go
- internal/formats/date_test.go
- internal/formats/time.go
- internal/formats/time_test.go
- internal/formats/duration.go
- internal/formats/duration_test.go
- internal/formats/email.go
- internal/formats/email_test.go
- internal/formats/idnemail.go
- internal/formats/idnemail_test.go
- internal/formats/hostname.go
- internal/formats/hostname_test.go
- internal/formats/idnhostname.go
- internal/formats/idnhostname_test.go
- internal/formats/ipv4.go
- internal/formats/ipv4_test.go
- internal/formats/ipv6.go
- internal/formats/ipv6_test.go
- internal/formats/uuid.go
- internal/formats/uuid_test.go
- internal/keywords/formatannotation/format.go
- internal/keywords/formatannotation/assert_test.go
- options.go
- internal/compile/compile.go
- internal/compile/metacheck.go
- formatassertion_test.go
- cmd/jsonschema/run.go
- cmd/jsonschema/formatassert_cli_test.go
- docs/formats.md

## Data / API impact
- DB impact: no
- API impact: yes: new library option and new CLI flag (additive, default off)
- Migration needed: no

## Constraints for implementation
- Use the standard library only; go.mod gets no require lines.
- Do not use net.ParseIP, netip.ParseAddr or net/mail.ParseAddress as the sole check. They accept zones, leading forms or address syntax beyond RFC 2673, RFC 4291 and RFC 5321. Hand-written parsers keep the strict behaviour.
- Format errors must never include the instance value.
- The compiled Schema stays immutable. The mode is fixed at compile time, with no per-call mutation.
- Touch the crowded files options.go, internal/compile/compile.go and cmd/jsonschema/run.go only to wire in the option, the flag and the boolean. Put logic in new files.
- Meta-schema validation never asserts format, whatever the caller's mode.
- Each format has its own file and its own test file. No central format list or switch.

## Risks
- The leap-second rule needs offset-adjusted UTC 23:59:60 handling, which is easy to get wrong for negative offsets.
- The RFC 5321 quoted-string and domain-literal grammar is intricate, and length limits (64 for the local part, 255 overall) must be applied after parsing.
- If story 190's flag is recorded per meta-schema rather than per resource, embedded resources with a different $schema need care.

## Breaking changes
- none

## Sensitivity
- security
