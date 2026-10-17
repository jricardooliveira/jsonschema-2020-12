# Pipeline record — #15: unevaluatedProperties and unevaluatedItems (2020-12 unevaluated vocabulary)

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #150 · **Status:** waiting_human · **Stage:** gate-release-human
- **Started:** 2026-10-17 00:39:58 UTC · **Stopped:** 2026-10-17 00:48:53 UTC
- **Story spend:** ~$3.11 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-release: `factory approve 150` to accept, or `factory reject 150 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-release**:
  > 🚀 RELEASE SIGN-OFF (Checkpoint 3) — NOT READY: 1 gap(s) in the evidence:
  >   • Security finding (medium) not resolved: F3 (medium): the boundary rule requires a test asserting that no error OR annotation output contains a sentinel instance value. The tests only scan units. The unevaluatedProperties annotation also carries the applied property names, which come from the document. ADR-014 rule 4 allows this exemption only for properties, patternProperties and additionalProperties. No test checks Annotations() for the sentinel, and unevaluatedProperties is not added to the documented exemption.; F4 (low): the fail-closed branches ('evaluated-property/item record unavailable') in unevaluatedproperties.go and unevaluateditems.go have no test. They cannot currently run, because evalNode's attachItems always sets ctx.Items. This is a note, not a live defect.
  >   ◦ Release-agent note: The 'not' keyword may leak annotations: if the 'not' keyword registers evaluated properties or items after its subschema fails, later unevaluated* keywords will incorrectly see them as covered. This is a correctness issue owned by the 'not' implementation (story 186 or later). Recommend a follow-up investigation.
  >   ◦ Release-agent note: Keyword location and absolute keyword location in error units are not explicitly tested in unevaluated*_test.go files; they are only checked indirectly through pass/fail outcome. Recommend a spot-check of actual error output from one failing case.
  >   ◦ Release-agent note: The fail-closed code path (returning error when ctx.Items does not implement the property/item record interface) is not exercised by any test. This would be triggered only by a bug in story 191's EvalContext. Recommend adding a mock test if paranoia is desired.
  >   → Approve to release accepting the gaps above as a known risk, or reject with what must change — it goes back to the coder, then the tester.

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
| 2026-10-17 00:48:53 | gate | gate-release | ⏸️ awaiting human sign-off |
| 2026-10-17 00:48:53 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
