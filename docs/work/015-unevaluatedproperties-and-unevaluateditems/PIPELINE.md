# Pipeline record — #15: unevaluatedProperties and unevaluatedItems (2020-12 unevaluated vocabulary)

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #150 · **Status:** waiting_human · **Stage:** gate-release-human
- **Started:** 2026-10-17 00:39:58 UTC · **Stopped:** 2026-10-17 00:56:00 UTC
- **Story spend:** ~$5.47 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-release: `factory approve 150` to accept, or `factory reject 150 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-release**:
  > 🚀 RELEASE SIGN-OFF (Checkpoint 3) — READY: every evidence bar is met.
  >   ◦ Release-agent note: These keywords depend on story 191 to provide annotation tracking (EvaluatedItems, mark/rollback) and story 191 dropping annotations from failed subschemas. The diff uses the interfaces correctly, but if 191 does not roll back on subschema failure, unevaluated* will over-report covered members. Confirm story 191 is complete and has rollback implemented.
  >   ◦ Release-agent note: anyOf and oneOf must evaluate all branches to collect complete annotations (per core 10.3). If story 185 short-circuits after the first passing branch, later branches' annotations will be missing. Verify story 185 does not short-circuit.
  >   ◦ Release-agent note: The $dynamicRef test cases depend on story 188 propagating annotations from dynamically resolved targets; confirm story 188 is in place.
  >   ◦ Release-agent note: QA noted three low findings (out-of-range mark handling, caps testing gaps, linear scan performance): these are not blocking, but reviewers should know these improvements could be made in follow-up stories (mark validation stricter, output limits tested, indexed lookup on deep nesting).
  >   ◦ Security note (low): Fail-closed rule (low): both keywords return an error when the evaluated-location record is missing. But EvaluatedItems.since() returns nil when the mark is out of range (internal/eval/eval.go), so an inconsistent Scope is silently read as 'nothing evaluated'. That path never passes wrongly: the document is reported invalid, or the subschema is applied to every property or item. Still, the boundary rule says an inconsistent state should return an error rather than give a result.
  >   → Approve to release, or reject with what must change — it goes back to the coder, then the tester.

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-17 00:39:58 | run | run started | • |
| 2026-10-17 00:39:58 | agent | spec-agent | ✅ pass |
| 2026-10-17 00:39:58 | gate | gate-1-spec | ✅ Story 'unevaluatedProperties and unevaluatedItems (2020-12 unevaluated vocabulary)' has 8 AC and 4 tasks |
| 2026-10-17 00:39:58 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-17 00:39:58 | agent | architect-agent | ✅ pass |
| 2026-10-17 00:39:59 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it touches sensitive areas (security) |
| 2026-10-17 00:40:12 | agent | boundary-agent | ✅ pass · 12.7s · $0.1493 |
| 2026-10-17 00:40:12 | gate | gate-2-architect | ✅ Architecture approved. 9 modules, 3 risks flagged |
| 2026-10-17 00:40:12 | boss | authorize coder-agent:T1 | ✅ gate-2-architect passed |
| 2026-10-17 00:40:55 | agent | coder-agent | ✅ complete · T1 · 43.4s · $0.3009 |
| 2026-10-17 00:41:14 | gate | gate-build | ✅ [T1] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 00:41:14 | boss | authorize coder-agent:T2 | ✅ gate-2-architect passed |
| 2026-10-17 00:42:15 | agent | coder-agent | ✅ complete · T2 · 60.7s · $0.3628 |
| 2026-10-17 00:42:34 | gate | gate-build | ✅ [T2] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 00:42:34 | boss | authorize coder-agent:T3 | ✅ gate-2-architect passed |
| 2026-10-17 00:43:23 | agent | coder-agent | ✅ complete · T3 · 48.6s · $0.3316 |
| 2026-10-17 00:43:31 | gate | gate-build | ❌ [T3] go_build:pass, go_vet:fail, go_test:fail |
| 2026-10-17 00:43:31 | boss | authorize coder-agent:T3 | ✅ gate-2-architect passed |
| 2026-10-17 00:45:02 | agent | coder-agent | ✅ complete · T3 · 91.1s · $0.7218 |
| 2026-10-17 00:45:18 | gate | gate-build | ✅ [T3] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 00:45:19 | boss | authorize coder-agent:T4 | ✅ gate-2-architect passed |
| 2026-10-17 00:46:35 | agent | coder-agent | ✅ complete · T4 · 76.1s · $0.4505 |
| 2026-10-17 00:46:42 | gate | gate-build | ✅ [T4] go_test:pass, go_race:pass |
| 2026-10-17 00:46:42 | boss | authorize tester-agent | ✅ gate-build passed with all 4 task(s) implemented |
| 2026-10-17 00:48:11 | agent | tester-agent | ✅ warn · 89.4s · $0.7251 |
| 2026-10-17 00:48:11 | gate | gate-test | ✅ QA/security/performance passed (with warnings). AC covered: 8; findings: 6 new, 0 persistent, 0 resolved |
| 2026-10-17 00:48:12 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-17 00:48:53 | agent | release-agent | ✅ warn · 40.9s · $0.0707 |
| 2026-10-17 00:48:53 | gate | gate-release | ❌ REJECTED: Functional defect, verified by the operator on this candidate: schema {"allOf":[{"properties":{"a":{}}},{"unevaluatedProperties":false}]} with document {"a":1} returns valid; 2020-12 requires invalid, becaus… |
| 2026-10-17 00:49:34 | boss | authorize coder-agent:remediation | ✅ the operator's rejection at Checkpoint 3 (1 finding(s)) |
| 2026-10-17 00:53:46 | agent | coder-agent | ✅ complete · remediation · 252.1s · $1.6169 |
| 2026-10-17 00:53:54 | gate | gate-build | ✅ [remediation] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 00:53:55 | boss | authorize tester-agent | ✅ gate-build passed with all 4 task(s) implemented |
| 2026-10-17 00:54:56 | agent | tester-agent | ✅ pass · 61.5s · $0.6473 |
| 2026-10-17 00:54:56 | gate | gate-test | ✅ QA/security/performance passed. AC covered: 8; findings: 3 new, 0 persistent, 0 resolved |
| 2026-10-17 00:54:56 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-17 00:55:59 | agent | release-agent | ✅ pass · 62.8s · $0.0900 |
| 2026-10-17 00:56:00 | gate | gate-release | ⏸️ awaiting human sign-off |
| 2026-10-17 00:56:00 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
