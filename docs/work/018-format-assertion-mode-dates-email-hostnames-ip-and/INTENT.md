# INTENT #18

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Add the optional format-assertion mode. It is enabled by an opt-in library option, by a CLI flag, or when a meta-schema declares the format-assertion vocabulary as required. In this mode, 'format' fails validation when the value does not match. Default behaviour stays annotation-only, and unknown formats are never an error. Implement these format checks from the 2020-12 validation spec, as far as practical, using only the standard library: date-time, date, time, duration, email, idn-email, hostname, idn-hostname, ipv4, ipv6 and uuid. Give format checks one registration point so each check lives in its own file. idn-hostname and idn-email get basic checks only, and the code records what they miss. Write positive and negative tests for each format and for each way of turning the mode on.

## Operator clarifications
- When a value looks almost right but breaks a fine point of the official standard, should it fail in format-checking mode? For example: an IP address with a leading zero (192.168.01.1), a time with second 60 that is not a real leap second, or an IPv6 address with a network-card suffix (fe80::1%eth0). — Follow the standards strictly — Recommended. Each format follows its official definition (RFC 3339 for dates and times, RFC 2673 for IPv4, RFC 4291 for IPv6, RFC 1123 for hostnames, RFC 4122 for uuid). All three examples fail. This is what independent evaluators expect.
- Email addresses have rare but legal forms, such as a quoted name ("john doe"@example.com) or an address at a raw IP ([192.168.0.1]@... style, written user@[192.168.0.1]). How complete should the 'email' check be? — Accept every form the email standard allows — Recommended. Follows the full address rules of RFC 5321, including quoted names and IP addresses in brackets. Nothing legal is wrongly rejected.
- A meta-schema can list the format-checking rules as 'optional' instead of 'required'. The spec lets a validator choose what to do then. If a schema's meta-schema lists them as optional and the caller did not switch the mode on, what should happen? — Stay annotation-only — Recommended. Formats are only checked when the caller turns the mode on or the meta-schema makes it required. Keeps the default predictable: nothing fails unless someone clearly asked for it.
- This story checks 11 formats. Others listed in the brief, such as uri, regex and json-pointer, come in a later story. Until then, what should format-checking mode do when a schema uses one of those not-yet-built formats? — Let it pass, and say so in the docs — Recommended. These formats are treated like unknown formats, which are never an error. The docs list which formats are checked so far. The later story turns each one on.
- Some rules leave small details open. Should these forms count as valid? A lowercase 't' or 'z' in a date-time (2024-01-01t10:00:00z), and a hostname that ends with a dot (example.com.). — Lowercase t/z is valid, a trailing dot is not — Recommended. The date-and-time standard (RFC 3339) explicitly allows lowercase. The hostname standard does not include the final dot, and most checkers reject it.
- International hostnames and email addresses (idn-hostname, idn-email) only get basic checks, because the full Unicode tables are not in Go's standard library. Which basic checks should they get? — Structure plus simple character rules — Recommended. Checks the overall length, the length of each part between the dots, that no part starts or ends with a hyphen, and that there are no spaces, control characters or stray punctuation. Plain-ASCII hostnames follow the same rules as 'hostname'. Anything that needs the Unicode tables is accepted and listed in the docs as not checked.
- A uuid is 32 letters and digits in groups, like 2eb8aa08-aa98-11ea-b4aa-73b441d16380. One digit inside it says which kind of uuid it is (versions 1 to 8 today), and another marks its family. If the groups and characters are right but that kind digit is unusual (for example 0 or 9), should it pass? — Check only the shape — Recommended. Accept any 8-4-4-4-12 group of hex digits, in upper or lower case, whatever the kind digit says. Matches what the 2020-12 spec's format definition asks for, and future uuid kinds never fail.
- A duration is written like P3D (3 days) or PT1H30M (1 hour 30 minutes). The standard allows weeks only on their own (P2W), so a mix like P1W3D breaks its rules even though some tools write it. Should a week-plus-days mix fail? — Fail, follow the standard — Recommended. In line with your answer to keep to the standards strictly (RFC 3339 appendix A). P2W passes, P1W3D fails, and so do empty forms like P or PT.
- Operator note on refining again — Add to this story (routed from stories 189 and 190): (1) with format-assertion mode ON, the meta-schema check of a schema must still never assert format: test that a schema with "$id": "http://example.com/a b" (violates uri-reference) still compiles with the mode on and off; (2) read story 190's recorded format-assertion vocabulary flag (required in the meta-schema's $vocabulary turns the mode on; optional does not) instead of re-parsing $vocabulary. Keep everything else in the accepted plan.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The format keyword's compile function needs the effective assert boolean. If the compile context type passed to keyword compile functions is defined in internal/keywords/registry.go (a crowded file) rather than internal/compile, one field or accessor must be added there. That makes a 33rd file and goes over the 32-file limit. If so, move T6 (docs/formats.md) into story 197, which already documents partly checked formats, instead of adding the file here.
- options.go is assumed to hold the library Options. If options are functional, the new option could live in a new root file instead, which would remove the crowded-file write.
- The exact name and location of story 190's recorded format-assertion vocabulary flag (expected in internal/compile/vocabulary.go or vocabulary_apply.go) has not been verified. This proposal only reads it.
- It is unclear whether internal/compile/metacheck.go already isolates meta-schema evaluation from caller options. If it does, the write claim there reduces to a test-only concern, which is covered in formatassertion_test.go.
- The convention prefers testdata/<area>/<keyword>.json cases. Adding 11 testdata files would go over the file limit, so format tests are Go table tests in each format's own _test.go file, with RFC/spec citations.
- The CLI flag is assumed to be registered in cmd/jsonschema/run.go, because main.go only wires to run. This was not verified against the truncated inventory.

## Proposed outcome

See `SPEC.md` in this directory — the spec-agent's acceptance criteria are
the machine-readable statement of the outcome, derived from the problem above.

## Affected users and systems

See `PLAN.md` → *Files that change*, and the project's `project-spec.json`
for the systems in scope.

## Constraints

The project's `PROJECT_RULES.md` and `project-spec.json` (stack, conventions,
NFRs and the `forbidden` list) are injected into every agent and bind this work.

## Open questions

Recorded in `SPEC.md` → *Open questions*. A non-empty list parks the run at
Checkpoint 1 for the operator rather than failing it.
