# Pipeline record — US-0171: Rename the Go module path to github.com/jricardooliveira/jsonschema-2020-12

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #171 · **Status:** waiting_human · **Stage:** gate-2-human
- **Started:** 2026-10-18 10:13:26 UTC · **Stopped:** 2026-10-18 10:14:47 UTC
- **Story spend:** ~$0.86 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-2-architect: `factory approve 171` to accept, or `factory reject 171 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-2-architect**:
  > ⚠️ BREAKING CHANGES detected — these will affect existing API consumers:
  >   • The Go module import path changes from github.com/neoface/jsonschema (and its /internal/... and /cmd/jsonschema subpaths) to github.com/jricardooliveira/jsonschema-2020-12. Existing library consumers must update their imports and go get target. The package name and API are unchanged.
  >   • The Go import path changes from github.com/neoface/jsonschema (and its /internal/... and /cmd/jsonschema subpaths) to github.com/jricardooliveira/jsonschema-2020-12. Existing library consumers must update their imports and their go get target. The package name `jsonschema`, the exported API, the CLI flags, the output and the exit codes are unchanged. The operator has already accepted this break and chose to add no migration note.
  >   → Approve these breaking changes?

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-18 10:13:26 | run | run started | • |
| 2026-10-18 10:13:46 | agent | spec-agent | ✅ pass · 19.6s · $0.2732 |
| 2026-10-18 10:13:47 | gate | gate-1-spec | ✅ Story 'Rename the Go module path to github.com/jricardooliveira/jsonschema-2020-12' has 7 AC and 1 tasks |
| 2026-10-18 10:13:47 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-18 10:14:36 | agent | architect-agent | ✅ pass · 49.3s · $0.4646 |
| 2026-10-18 10:14:37 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes |
| 2026-10-18 10:14:47 | agent | boundary-agent | ✅ pass · 9.9s · $0.1265 |
| 2026-10-18 10:14:47 | gate | gate-2-architect | ⏸️ awaiting human sign-off |
| 2026-10-18 10:14:47 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
