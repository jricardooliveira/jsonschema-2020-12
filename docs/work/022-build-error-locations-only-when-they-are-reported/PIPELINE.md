# Pipeline record — #22: Build keyword locations only when a report or Result shows them

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #157 · **Status:** waiting_human · **Stage:** gate-release-human
- **Started:** 2026-10-17 10:53:58 UTC · **Stopped:** 2026-10-17 11:25:18 UTC
- **Story spend:** ~$3.93 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-release: `factory approve 157` to accept, or `factory reject 157 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-release**:
  > 🚀 RELEASE SIGN-OFF (Checkpoint 3) — READY: every evidence bar is met.
  >   ◦ Release-agent note: Dead code: location.go defines nodeLoc and kwLoc types that are never used. The actual code uses a simpler lazyLoc type in eval.go. Remove the unused types from location.go before shipping, or refactor to use the location.go types and remove the lazyLoc from eval.go.
  >   ◦ Release-agent note: Test/code mismatch: location_test.go tests nodeLoc and kwLoc against an eager reimplementation, but these types are not deployed. The actual lazyLoc type in eval.go has no dedicated unit tests; its correctness is verified only by end-to-end tests. Add unit tests for lazyLoc or refactor to use the location.go types consistently.
  >   ◦ Release-agent note: Hard-coded limit edge values in limits_edge_regression_test.go (depth 4, ref-chain 5, steps 186, output-units 4, report-bytes 414, annotations 37) are not derived from the code; the acceptance criterion says limits should be read from the code, not repeated as numbers. Consider reading these from options.go or limitflags.go instead.
  >   ◦ Release-agent note: The change touches internal/keywords files in multiple vocabularies (applicator, unevaluated) to wire the lazy location interface. All touched files are in the granted scope, but the if.go modification introduces a branchLoc type to carry the base absolute location; confirm this works correctly in verbose's second evaluation pass.
  >   ◦ Security note (none): No new security issue. Locations are still built only from schema-static data (pathFrame dynamic/static, Resource, Item.Location) and never from instance values. KeywordLoc() returns "" when ctx.Loc is nil, which matches the old zero-value KeywordLocation, so it is not a new way to fail open. The limit error paths are unchanged, and the new edge test pins them to exit 2.
  >   → Approve to release, or reject with what must change — it goes back to the coder, then the tester.

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
| 2026-10-17 11:03:58 | gate | gate-question | ❌ REJECTED: yes: all nine files granted for this run (no other run is in flight on them). Limit the edits to carrying the lazy location reference instead of the strings: replace the KeywordLocation/AbsoluteKeywordLocati… |
| 2026-10-17 11:14:22 | boss | authorize coder-agent:T3 | ✅ gate-2-architect passed |
| 2026-10-17 11:16:30 | agent | coder-agent | ✅ complete · T3 · 128.3s · $0.6638 |
| 2026-10-17 11:18:08 | gate | gate-build | ✅ [T3] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 11:18:09 | boss | authorize coder-agent:T5 | ✅ gate-2-architect passed |
| 2026-10-17 11:18:42 | agent | coder-agent | ✅ complete · T5 · 33.2s · $0.3088 |
| 2026-10-17 11:18:43 | gate | gate-build | ✅ [T5] go_test:pass, go_race:pass |
| 2026-10-17 11:18:43 | boss | authorize coder-agent:T6 | ✅ gate-2-architect passed |
| 2026-10-17 11:23:19 | agent | coder-agent | ✅ complete · T6 · 275.9s · $0.4504 |
| 2026-10-17 11:23:40 | gate | gate-build | ✅ [T6] go_test:pass, go_race:pass |
| 2026-10-17 11:23:40 | boss | authorize tester-agent | ✅ gate-build passed with all 6 task(s) implemented |
| 2026-10-17 11:24:42 | agent | tester-agent | ✅ warn · 61.1s · $0.6991 |
| 2026-10-17 11:24:42 | gate | gate-test | ✅ QA/security/performance passed (with warnings). AC covered: 8; findings: 3 new, 0 persistent, 0 resolved |
| 2026-10-17 11:24:42 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-17 11:25:17 | agent | release-agent | ✅ warn · 35.0s · $0.0695 |
| 2026-10-17 11:25:18 | gate | gate-release | ⏸️ awaiting human sign-off |
| 2026-10-17 11:25:18 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
