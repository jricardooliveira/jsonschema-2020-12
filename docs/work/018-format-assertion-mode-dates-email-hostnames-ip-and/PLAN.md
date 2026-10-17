# PLAN #18: Format assertion: mode, dates, email, hostnames, IP and UUID

- Date: 2026-10-17
- Status: passed gate 2 with nothing for the operator to decide (run #153, 2026-10-17 02:32 UTC)
- Design: see `ADR-018-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/formats/registry.go`
- `options.go`
- `internal/compile/compile.go`
- `internal/compile/metacheck.go`
- `internal/keywords/formatannotation/format.go`
- `internal/keywords/formatannotation/assert_test.go`
- `internal/formats/datetime.go`
- `internal/formats/datetime_test.go`
- `internal/formats/date.go`
- `internal/formats/date_test.go`
- `internal/formats/time.go`
- `internal/formats/time_test.go`
- `internal/formats/duration.go`
- `internal/formats/duration_test.go`
- `internal/formats/hostname.go`
- `internal/formats/hostname_test.go`
- `internal/formats/idnhostname.go`
- `internal/formats/idnhostname_test.go`
- `internal/formats/email.go`
- `internal/formats/email_test.go`
- `internal/formats/idnemail.go`
- `internal/formats/idnemail_test.go`
- `internal/formats/ipv4.go`
- `internal/formats/ipv4_test.go`
- `internal/formats/ipv6.go`
- `internal/formats/ipv6_test.go`
- `internal/formats/uuid.go`
- `internal/formats/uuid_test.go`
- `cmd/jsonschema/run.go`
- `cmd/jsonschema/formatassert_cli_test.go`
- `formatassertion_test.go`
- `docs/formats.md`

Modules the architecture expects to touch:

- `internal/formats/registry.go`
- `internal/formats/datetime.go`
- `internal/formats/datetime_test.go`
- `internal/formats/date.go`
- `internal/formats/date_test.go`
- `internal/formats/time.go`
- `internal/formats/time_test.go`
- `internal/formats/duration.go`
- `internal/formats/duration_test.go`
- `internal/formats/email.go`
- `internal/formats/email_test.go`
- `internal/formats/idnemail.go`
- `internal/formats/idnemail_test.go`
- `internal/formats/hostname.go`
- `internal/formats/hostname_test.go`
- `internal/formats/idnhostname.go`
- `internal/formats/idnhostname_test.go`
- `internal/formats/ipv4.go`
- `internal/formats/ipv4_test.go`
- `internal/formats/ipv6.go`
- `internal/formats/ipv6_test.go`
- `internal/formats/uuid.go`
- `internal/formats/uuid_test.go`
- `internal/keywords/formatannotation/format.go`
- `internal/keywords/formatannotation/assert_test.go`
- `options.go`
- `internal/compile/compile.go`
- `internal/compile/metacheck.go`
- `formatassertion_test.go`
- `cmd/jsonschema/run.go`
- `cmd/jsonschema/formatassert_cli_test.go`
- `docs/formats.md`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Format registry and assertion-mode plumbing
2. **T2** — Date and time formats  _(after T1)_
3. **T4** — IP address and uuid formats  _(after T1)_
4. **T3** — Hostname and email formats, ASCII and international  _(after T1, T4)_
5. **T5** — CLI flag and end-to-end enablement tests  _(after T1, T2, T3, T4)_
6. **T6** — Document checked and partly checked formats  _(after T3)_

## Risks

- The leap-second rule needs offset-adjusted UTC 23:59:60 handling, which is easy to get wrong for negative offsets.
- The RFC 5321 quoted-string and domain-literal grammar is intricate, and length limits (64 for the local part, 255 overall) must be applied after parsing.
- If story 190's flag is recorded per meta-schema rather than per resource, embedded resources with a different $schema need care.

Implementation constraints the coder must honor:

- Use the standard library only; go.mod gets no require lines.
- Do not use net.ParseIP, netip.ParseAddr or net/mail.ParseAddress as the sole check. They accept zones, leading forms or address syntax beyond RFC 2673, RFC 4291 and RFC 5321. Hand-written parsers keep the strict behaviour.
- Format errors must never include the instance value.
- The compiled Schema stays immutable. The mode is fixed at compile time, with no per-call mutation.
- Touch the crowded files options.go, internal/compile/compile.go and cmd/jsonschema/run.go only to wire in the option, the flag and the boolean. Put logic in new files.
- Meta-schema validation never asserts format, whatever the caller's mode.
- Each format has its own file and its own test file. No central format list or switch.

- DB impact: no · API impact: yes: new library option and new CLI flag (additive, default off) · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: assert_test.go shows four things: a registered stub format fails in the mode and passes outside it; unknown formats pass; non-strings pass; and the meta-check compiles a schema with mode on.
- **T2**: The table tests show these results: leap years and lowercase t/z pass; a non-leap second 60, P1W3D, P and PT fail.
- **T4**: The table tests show these results: 192.168.01.1 and fe80::1%eth0 fail; an embedded IPv4 tail passes; an upper-case uuid with version digit 0 or 9 passes.
- **T3**: The table tests show these results: a trailing dot fails; quoted local parts and IP domain literals pass; hyphen-edged labels fail; ASCII idn input matches the hostname results.
- **T5**: The CLI tests show exit 1 with the flag and exit 0 without it. The library tests cover the option, a required vocabulary, an optional vocabulary, not-yet-built formats, and the '$id' 'a b' case with mode on and off. All pass under -race.
- **T6**: docs/formats.md lists all 11 checked formats, the idn gaps and the 8 deferred formats.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
