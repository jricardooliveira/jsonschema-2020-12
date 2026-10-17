# Pipeline record — #22: Build keyword locations only when a report or Result shows them

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #157 · **Status:** waiting_human · **Stage:** gate-question
- **Started:** 2026-10-17 10:53:58 UTC · **Stopped:** 2026-10-17 11:03:58 UTC
- **Story spend:** ~$1.74 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-question: `factory approve 157` to accept, or `factory reject 157 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-question**:
  > The coder asks for files the approved design does not name: The design says only internal/eval changes. But keyword evaluate functions and EvalContext (internal/keywords/registry.go) take and copy location strings through the public fields KeywordLocation and AbsoluteKeywordLocation. Making those lazy needs the field replaced by a lazy reference with a render method. That means editing internal/keywords/registry.go and internal/keywords/annotate.go. It also means editing the keyword files that read ctx.KeywordLocation: applicator/contains.go, if.go, items.go, properties.go, prefixitems.go, and unevaluated/unevaluatedproperties.go and unevaluateditems.go. Not granted automatically: internal/keywords/registry.go (not in the approved design and not a test); internal/keywords/annotate.go (not in the approved design and not a test); internal/keywords/applicator/contains.go (not in the approved design and not a test); internal/keywords/applicator/if.go (not in the approved design and not a test); internal/keywords/applicator/items.go (not in the approved design and not a test); internal/keywords/applicator/properties.go (not in the approved design and not a test); internal/keywords/applicator/prefixitems.go (not in the approved design and not a test); internal/keywords/unevaluated/unevaluatedproperties.go (not in the approved design and not a test); internal/keywords/unevaluated/unevaluateditems.go (not in the approved design and not a test). An answer that starts with "yes" (a note may follow) approves these files for this run: the same pass runs again and may change them. An answer that starts with "no" (or a rejection) refuses them: the pass runs again without them, with your words as a finding.
  > paths: internal/keywords/registry.go, internal/keywords/annotate.go, internal/keywords/applicator/contains.go, internal/keywords/applicator/if.go, internal/keywords/applicator/items.go, internal/keywords/applicator/properties.go, internal/keywords/applicator/prefixitems.go, internal/keywords/unevaluated/unevaluatedproperties.go, internal/keywords/unevaluated/unevaluateditems.go

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-17 10:53:58 | run | run started | • |
| 2026-10-17 10:53:58 | agent | spec-agent | ✅ pass |
| 2026-10-17 10:53:58 | gate | gate-1-spec | ✅ Story 'Build keyword locations only when a report or Result shows them' has 8 AC and 6 tasks |
| 2026-10-17 10:53:59 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-17 10:53:59 | agent | architect-agent | ✅ pass |
| 2026-10-17 10:53:59 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it touches sensitive areas (security) |
| 2026-10-17 10:54:06 | agent | boundary-agent | ✅ pass · 7.1s · $0.1716 |
| 2026-10-17 10:54:06 | gate | gate-2-architect | ✅ Architecture approved. 7 modules, 3 risks flagged |
| 2026-10-17 10:54:07 | boss | authorize coder-agent:T1 | ✅ gate-2-architect passed |
| 2026-10-17 10:59:13 | agent | coder-agent | ✅ complete · T1 · 306.2s · $0.4709 |
| 2026-10-17 10:59:35 | gate | gate-build | ✅ [T1] go_test:pass, go_race:pass |
| 2026-10-17 10:59:35 | boss | authorize coder-agent:T2 | ✅ gate-2-architect passed |
| 2026-10-17 11:00:15 | agent | coder-agent | ✅ complete · T2 · 39.3s · $0.3276 |
| 2026-10-17 11:01:54 | gate | gate-build | ✅ [T2] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 11:01:54 | boss | authorize coder-agent:T4 | ✅ gate-2-architect passed |
| 2026-10-17 11:03:12 | agent | coder-agent | ✅ complete · T4 · 77.2s · $0.4605 |
| 2026-10-17 11:03:35 | gate | gate-build | ✅ [T4] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 11:03:35 | boss | authorize coder-agent:T3 | ✅ gate-2-architect passed |
| 2026-10-17 11:03:58 | agent | coder-agent | ✅ scope-request · T3 · 23.0s · $0.3084 |
| 2026-10-17 11:03:58 | gate | gate-question | ⏸️ awaiting human sign-off |
| 2026-10-17 11:03:58 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
