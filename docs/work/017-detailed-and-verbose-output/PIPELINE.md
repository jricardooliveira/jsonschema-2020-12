# Pipeline record — #17: Detailed and verbose output formats, with URI-safe absoluteKeywordLocation

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #149 · **Status:** waiting_human · **Stage:** gate-2-human
- **Started:** 2026-10-17 00:39:55 UTC · **Stopped:** 2026-10-17 00:40:19 UTC
- **Story spend:** ~$0.17 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-2-architect: `factory approve 149` to accept, or `factory reject 149 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-2-architect**:
  > ⚠️ BREAKING CHANGES detected — these will affect existing API consumers:
  >   • absoluteKeywordLocation for schema keys with spaces, '%' or non-ASCII characters is now percent-encoded in flag/basic output too
  >   • absoluteKeywordLocation fragments for schema keys with a space, '%', non-ASCII or other characters not allowed in a URI fragment are now percent-encoded in every output shape, including basic. Output for keys made only of allowed characters is unchanged.
  >   → Approve these breaking changes?
  >
  > 🧱 BOUNDARY REVIEW WARNINGS — tenant / authorization / API contract / security concerns found before any code is written:
  >   • api_contract: The design and acceptance criterion 6 say the existing library option and CLI flag raise DefaultMaxReportBytes, and the design claims cmd/jsonschema/run.go is read-only. ADR #14 says the CLI has no report-bytes flag in story 191 and that story 193 adds it. If the flag does not exist, a CLI user cannot raise the cap for 'verbose', and the T5 test of 'raise the cap' through the CLI cannot be written without editing run.go. The operator should settle whether this story adds the flag or only the library option is claimed here.
  >   • api_contract: The breaking change to absoluteKeywordLocation (percent-encoded fragment) is declared and justified because the value must be a valid URI. It applies to flag, basic, detailed and verbose alike. Consumers that compare the old unencoded string will see a difference only for keys with a space, '%', non-ASCII or other characters not allowed in a URI fragment.
  >   • security: Verbose needs passing units that the evaluator may not record today. If internal/eval/units.go starts keeping passing units on every Validate call, memory grows for hostile or large documents on all shapes, including flag, before the report byte cap is ever checked, because the cap only applies at render time. Counting those units against the existing MaxOutputUnits could instead turn a valid result into a limit error (exit 2) for flag and basic. The design does not say which of the two applies, and an operator should settle it.
  >   • security: Property names listed in annotations (properties, patternProperties, additionalProperties, unevaluatedProperties) can be private. The operator has decided to show names and never values, so this is accepted. It is noted here so the rule below is enforced in tests.
  >   → Approve the design with these boundary risks, or reject with the fix?

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-17 00:39:55 | run | run started | • |
| 2026-10-17 00:39:55 | agent | spec-agent | ✅ pass |
| 2026-10-17 00:39:55 | gate | gate-1-spec | ✅ Story 'Detailed and verbose output formats, with URI-safe absoluteKeywordLocation' has 8 AC and 5 tasks |
| 2026-10-17 00:39:56 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-17 00:39:56 | agent | architect-agent | ✅ pass |
| 2026-10-17 00:39:56 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security, pii) |
| 2026-10-17 00:40:19 | agent | boundary-agent | ✅ warn · 22.7s · $0.1712 |
| 2026-10-17 00:40:19 | gate | gate-2-architect | ⏸️ awaiting human sign-off |
| 2026-10-17 00:40:19 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
