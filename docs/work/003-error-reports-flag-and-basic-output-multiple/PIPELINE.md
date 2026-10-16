# Pipeline record — #3: Error reports: flag and basic output, multiple documents

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #137 · **Status:** waiting_human · **Stage:** gate-1-human
- **Started:** 2026-10-16 19:00:55 UTC · **Stopped:** 2026-10-16 19:00:55 UTC
- **Story spend:** ~$0.00 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-1-spec: `factory approve 137` to accept, or `factory reject 137 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-1-spec**:
  > 📐 UNAUTHORISED THRESHOLD — your request used a term that only means something once a number is attached, and never gave the number. These criteria either leave it undefined or fill it in with a value you never approved:
  >   • 'short' — still undefined in: In 'basic', a valid document gives exactly {"valid": true}. No annotations are included in this story. An invalid document gives {"valid": false, "errors": [...]}. Each error unit has 'instanceLocation' and 'keywordLocation' as JSON Pointers, 'absoluteKeywordLocation' only where an absolute URI is known, and 'error', a short sentence naming the failed rule.
  >   → Reject with the definition (e.g. "overdue = HIGH, not CLOSED, created >24h ago"), or approve to accept whatever the agents choose.

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-16 19:00:55 | run | run started | • |
| 2026-10-16 19:00:55 | agent | spec-agent | ✅ pass |
| 2026-10-16 19:00:55 | gate | gate-1-spec | ⏸️ awaiting human sign-off |
| 2026-10-16 19:00:55 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
