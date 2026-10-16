# Pipeline record — #8: Applicator keywords allOf, anyOf, oneOf, not and if/then/else, with keyword locations rooted at the schema root

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #142 · **Status:** waiting_human · **Stage:** gate-release-human
- **Started:** 2026-10-16 20:40:08 UTC · **Stopped:** 2026-10-16 20:47:11 UTC
- **Story spend:** ~$3.23 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-release: `factory approve 142` to accept, or `factory reject 142 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-release**:
  > 🚀 RELEASE SIGN-OFF (Checkpoint 3) — NOT READY: 1 gap(s) in the evidence:
  >   • Security finding (medium) not resolved: F1 (medium, hostile-input DoS): compileBranch in internal/keywords/applicator/if.go compiles the sibling 'if' subschema again inside both then.go and else.go, and evalBranch evaluates it again for each branch. A schema that nests if/then/else inside 'if' therefore doubles the compile and evaluation work at every level. A few KB of nesting (depth around 30) means about 2^30 compiled nodes, so a hostile schema can hang Compile. Compile is not covered by the future evaluation-step limit, and the nesting-depth limit from story 193 does not stop this at generous defaults. Safe default: compile 'if' once and have the branches use that single compiled node, so the work grows linearly.
  >   ◦ Release-agent note: F1 (medium): then/else recompile and re-evaluate sibling 'if' instead of compiling once. Nested conditionals in a hostile schema can therefore blow up compile and evaluation time exponentially. The 'if' subschema gets /then or /else locations instead of /if. Tester recommends fixing before stories 191/192 build on it, which do not block this release.
  >   → Approve to release accepting the gaps above as a known risk, or reject with what must change — it goes back to the coder, then the tester.

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-16 20:40:08 | run | run started | • |
| 2026-10-16 20:40:08 | agent | spec-agent | ✅ pass |
| 2026-10-16 20:40:08 | gate | gate-1-spec | ✅ Story 'Applicator keywords allOf, anyOf, oneOf, not and if/then/else, with keyword locations rooted at the schema root' has 10 AC and 5 tasks |
| 2026-10-16 20:40:09 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-16 20:40:09 | agent | architect-agent | ✅ pass |
| 2026-10-16 20:40:09 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it declares breaking changes; it touches sensitive areas (security) |
| 2026-10-16 20:40:24 | agent | boundary-agent | ✅ pass · 14.7s · $0.1619 |
| 2026-10-16 20:40:24 | gate | gate-2-architect | ✅ APPROVED: Approved by human reviewer |
| 2026-10-16 20:40:41 | boss | authorize coder-agent:T1 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 20:41:23 | agent | coder-agent | ✅ complete · T1 · 41.8s · $0.3077 |
| 2026-10-16 20:41:24 | gate | gate-build | ❌ [T1] go_build:pass, go_vet:pass, go_test:pass, go_race:pass, task_tests:fail |
| 2026-10-16 20:41:25 | boss | authorize coder-agent:T1 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 20:42:02 | agent | coder-agent | ✅ complete · T1 · 36.9s · $0.6348 |
| 2026-10-16 20:42:03 | gate | gate-build | ✅ [T1] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 20:42:03 | boss | authorize coder-agent:T2 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 20:42:48 | agent | coder-agent | ✅ complete · T2 · 45.1s · $0.3319 |
| 2026-10-16 20:42:53 | gate | gate-build | ✅ [T2] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 20:42:53 | boss | authorize coder-agent:T3 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 20:43:30 | agent | coder-agent | ✅ complete · T3 · 36.8s · $0.3054 |
| 2026-10-16 20:43:35 | gate | gate-build | ✅ [T3] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 20:43:36 | boss | authorize coder-agent:T4 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 20:44:34 | agent | coder-agent | ✅ complete · T4 · 57.6s · $0.3437 |
| 2026-10-16 20:44:41 | gate | gate-build | ✅ [T4] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 20:44:41 | boss | authorize coder-agent:T5 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 20:45:27 | agent | coder-agent | ✅ complete · T5 · 45.9s · $0.3678 |
| 2026-10-16 20:45:30 | gate | gate-build | ✅ [T5] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 20:45:30 | boss | authorize tester-agent | ✅ gate-build passed with all 5 task(s) implemented |
| 2026-10-16 20:46:27 | agent | tester-agent | ✅ warn · 56.2s · $0.6930 |
| 2026-10-16 20:46:27 | gate | gate-test | ✅ QA/security/performance passed (with warnings). AC covered: 10; findings: 3 new, 0 persistent, 0 resolved |
| 2026-10-16 20:46:27 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-16 20:47:10 | agent | release-agent | ✅ warn · 42.8s · $0.0852 |
| 2026-10-16 20:47:11 | gate | gate-release | ⏸️ awaiting human sign-off |
| 2026-10-16 20:47:11 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
