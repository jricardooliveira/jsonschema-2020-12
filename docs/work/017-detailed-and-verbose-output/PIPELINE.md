# Pipeline record — #17: Detailed and verbose output formats, with URI-safe absoluteKeywordLocation

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #149 · **Status:** waiting_human · **Stage:** gate-2-human
- **Started:** 2026-10-17 00:39:55 UTC · **Stopped:** 2026-10-17 00:41:19 UTC
- **Story spend:** ~$0.76 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-2-architect: `factory approve 149` to accept, or `factory reject 149 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-2-architect**:
  > ⚠️ BREAKING CHANGES detected — these will affect existing API consumers:
  >   • absoluteKeywordLocation for schema keys that contain spaces, '%' or non-ASCII characters is now percent-encoded, in flag and basic output too (accepted by the reviewer)
  >   • absoluteKeywordLocation is now percent-encoded in its fragment, in flag, basic, detailed and verbose output alike. Only schema keys with a space, '%', non-ASCII or other characters not allowed in a URI fragment are affected. The operator accepted this.
  >   → Approve these breaking changes?
  >
  > 🧱 BOUNDARY REVIEW WARNINGS — tenant / authorization / API contract / security concerns found before any code is written:
  >   • security: The design still does not say how verbose gets its passing units. The shape is chosen at render time, after Validate has run, so internal/eval/units.go may have to record passing units on every Validate call. If it does, memory for hostile or large documents grows on flag and basic as well, before the report byte cap is checked, because that cap only applies while rendering. If those units count against MaxOutputUnits, a result that is valid today could become a limit error (exit 2) for flag and basic. The operator's Checkpoint 2 decision settled only the CLI flag question, so a human should decide which of the two applies. The rules below give the safe constraints either way.
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
| 2026-10-17 00:40:19 | gate | gate-2-architect | ❌ REJECTED: Keep the design; settle: this story uses only the existing library option for DefaultMaxReportBytes and does NOT add the CLI flag or edit cmd/jsonschema/run.go; the CLI flag and its 'raise the cap through th… |
| 2026-10-17 00:40:38 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-17 00:41:00 | agent | architect-agent | ✅ pass · 21.4s · $0.4151 |
| 2026-10-17 00:41:00 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security, pii) |
| 2026-10-17 00:41:19 | agent | boundary-agent | ✅ warn · 18.1s · $0.1768 |
| 2026-10-17 00:41:19 | gate | gate-2-architect | ⏸️ awaiting human sign-off |
| 2026-10-17 00:41:19 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
