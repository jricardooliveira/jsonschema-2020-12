# Pipeline record — #6: 2020-12 object keywords: properties, patternProperties, additionalProperties, propertyNames, dependentSchemas, maxProperties, minProperties, required, dependentRequired

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #141 · **Status:** waiting_human · **Stage:** gate-2-human
- **Started:** 2026-10-16 20:16:13 UTC · **Stopped:** 2026-10-16 20:19:11 UTC
- **Story spend:** ~$1.06 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-2-architect: `factory approve 141` to accept, or `factory reject 141 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-2-architect**:
  > 🧱 BOUNDARY REVIEW WARNINGS — tenant / authorization / API contract / security concerns found before any code is written:
  >   • api_contract: The operator decided: failures come out by sorted member name, then missing required/dependentRequired names in schema order. The design only delivers keyword order first, then sorted member name inside each keyword. Keywords run in sorted keyword-name order: additionalProperties, dependentRequired, dependentSchemas, maxProperties, minProperties, patternProperties, properties, propertyNames, required.
  >   • api_contract: Because of that order, dependentRequired's missing-name failures come BEFORE the per-member failures of properties. This contradicts the operator's 'members first, then dependentRequired missing names'. The design lists this in its risks but does not put it to the operator. Only 'required' matches the decision exactly.
  >   • api_contract: A member failing under additionalProperties is also reported before a lexically earlier member failing under properties or patternProperties. A global sort by member name across keywords does not hold either.
  >   • api_contract: The planned second test, 'unit order equals sorted names followed by missing names', would fail or be written too narrowly for any mixed case that includes additionalProperties or dependentRequired. A human should confirm that keyword order, then sorted member, then schema order is the accepted contract, or decide that the crowded compile.go/eval.go may be edited to reorder.
  >   → Approve the design with these boundary risks, or reject with the fix?

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-16 20:16:13 | run | run started | • |
| 2026-10-16 20:16:13 | agent | spec-agent | ✅ pass |
| 2026-10-16 20:16:13 | gate | gate-1-spec | ✅ Story '2020-12 object keywords: properties, patternProperties, additionalProperties, propertyNames, dependentSchemas, maxProperties, minProperties, required, dependentRequired' has 11 AC and 5 tasks |
| 2026-10-16 20:16:13 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-16 20:16:13 | agent | architect-agent | ✅ pass |
| 2026-10-16 20:16:14 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it touches sensitive areas (pii) |
| 2026-10-16 20:16:32 | agent | boundary-agent | ✅ warn · 18.6s · $0.1741 |
| 2026-10-16 20:16:32 | gate | gate-2-architect | ❌ REJECTED: Keep the design and settle the ordering point without touching the decoder: iterate object members in deterministic sorted key order (byte-wise UTF-8 order of the member names), not document order (the opera… |
| 2026-10-16 20:17:31 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-16 20:18:47 | agent | architect-agent | ✅ pass · 76.5s · $0.6877 |
| 2026-10-16 20:18:48 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it touches sensitive areas (pii) |
| 2026-10-16 20:19:10 | agent | boundary-agent | ✅ warn · 22.8s · $0.1972 |
| 2026-10-16 20:19:11 | gate | gate-2-architect | ⏸️ awaiting human sign-off |
| 2026-10-16 20:19:11 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
