# Pipeline record — #10: Controlled schema registry for external references

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #145 · **Status:** waiting_human · **Stage:** gate-2-human
- **Started:** 2026-10-16 21:54:26 UTC · **Stopped:** 2026-10-16 21:54:59 UTC
- **Story spend:** ~$0.19 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-2-architect: `factory approve 145` to accept, or `factory reject 145 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-2-architect**:
  > 🧱 BOUNDARY REVIEW WARNINGS — tenant / authorization / API contract / security concerns found before any code is written:
  >   • api_contract: The design does not say how the CLI tells `--ref <uri>=<path>` from `--ref <path>`. A path can contain '=' (./a=b.json), a Windows path can contain ':' (C:\x.json), and a URI can contain '=' in its query (https://x/a?v=1). A wrong guess maps a file to the wrong address or reads the wrong path. This is part of a public CLI contract that scripts will depend on, so a human should settle the rule before it ships. Proposed default: a value is a mapping only if the text before the first '=' parses as an absolute URI with a scheme of two or more characters and no whitespace. Otherwise the whole value is a path. Callers whose URI contains '=' or whose path looks like a URI can prefix the path with './'.
  >   → Approve the design with these boundary risks, or reject with the fix?

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-16 21:54:26 | run | run started | • |
| 2026-10-16 21:54:26 | agent | spec-agent | ✅ pass |
| 2026-10-16 21:54:26 | gate | gate-1-spec | ✅ Story 'Controlled schema registry for external references' has 12 AC and 5 tasks |
| 2026-10-16 21:54:27 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-16 21:54:27 | agent | architect-agent | ✅ pass |
| 2026-10-16 21:54:27 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it touches sensitive areas (security) |
| 2026-10-16 21:54:59 | agent | boundary-agent | ✅ warn · 32.1s · $0.1869 |
| 2026-10-16 21:54:59 | gate | gate-2-architect | ⏸️ awaiting human sign-off |
| 2026-10-16 21:54:59 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
