# Pipeline record — #20: README and docs/ that describe the finished validator, proven by tests

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #155 · **Status:** waiting_human · **Stage:** gate-1-human
- **Started:** 2026-10-17 05:21:08 UTC · **Stopped:** 2026-10-17 05:21:09 UTC
- **Story spend:** ~$0.00 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-1-spec: `factory approve 155` to accept, or `factory reject 155 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-1-spec**:
  > 📐 UNAUTHORISED THRESHOLD — your request used a term that only means something once a number is attached, and never gave the number. These criteria either leave it undefined or fill it in with a value you never approved:
  >   • 'short' — still undefined in: docs/library.md documents the load-once, check-many API (compile with registry, limits and format mode fixed at compile time, then validate many documents, safe for concurrent use) with a short example. The same example exists as a Go Example function in example_test.go, and `go test` checks its `// Output:`.
  >   → Reject with the definition (e.g. "overdue = HIGH, not CLOSED, created >24h ago"), or approve to accept whatever the agents choose.

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-17 05:21:08 | run | run started | • |
| 2026-10-17 05:21:09 | agent | spec-agent | ✅ pass |
| 2026-10-17 05:21:09 | gate | gate-1-spec | ⏸️ awaiting human sign-off |
| 2026-10-17 05:21:09 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
