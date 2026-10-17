# Pipeline record — #25: Faster schema loading: build the embedded 2020-12 meta-schema once and check schemas without annotations

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #160 · **Status:** waiting_human · **Stage:** gate-release-human
- **Started:** 2026-10-17 12:41:47 UTC · **Stopped:** 2026-10-17 12:58:10 UTC
- **Story spend:** ~$4.01 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-release: `factory approve 160` to accept, or `factory reject 160 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-release**:
  > 🚀 RELEASE SIGN-OFF (Checkpoint 3) — NOT READY: 1 gap(s) in the evidence:
  >   • 1 acceptance criteria were never assessed by the tester (§5.1): ['The shared built-in result and every compiled Schema are immutable after they are built and safe for concurrent use. A new test that runs many concurrent Compile and Validate calls from separate goroutines, including the very first Compile in the process, passes under go test -race.']
  >   ◦ Release-agent note: The tester gave QA a 'warn' verdict for three medium test gaps: (1) no test of a custom meta-schema using unevaluated* keywords that depends on annotations, (2) no explicit test when a caller registers a schema at a built-in URI, and (3) the concurrency test may not reliably race the first build. None of these gaps cause a result change (all existing tests pass unchanged), but they leave specific edge cases without explicit coverage. If a future issue arises, the fix is localized to metacheck.go or registry.go.
  >   ◦ Release-agent note: The benchmark improvement (70-82% on repeated Compile) applies only to repeated loads; the first Compile in a process still pays the full cost of building the meta-schemas, so a one-shot CLI run gets no benefit. The FirstLoadOrder benchmark is unchanged, which is correct.
  >   ◦ Release-agent note: The embedded-resource index is built twice per process (once in registry.go and once in compile.go). Both are read-only after build, causing no correctness issue, but the redundancy was accepted as simpler than unifying the two caches.
  >   → Approve to release accepting the gaps above as a known risk, or reject with what must change — it goes back to the coder, then the tester.

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-17 12:41:47 | run | run started | • |
| 2026-10-17 12:41:47 | agent | spec-agent | ✅ pass |
| 2026-10-17 12:41:47 | gate | gate-1-spec | ✅ Story 'Faster schema loading: build the embedded 2020-12 meta-schema once and check schemas without annotations' has 9 AC and 6 tasks |
| 2026-10-17 12:41:48 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-17 12:41:48 | agent | architect-agent | ✅ pass |
| 2026-10-17 12:41:48 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it touches sensitive areas (security) |
| 2026-10-17 12:42:04 | agent | boundary-agent | ✅ pass · 16.2s · $0.1842 |
| 2026-10-17 12:42:04 | gate | gate-2-architect | ✅ Architecture approved. 13 modules, 5 risks flagged |
| 2026-10-17 12:42:05 | boss | authorize coder-agent:T1 | ✅ gate-2-architect passed |
| 2026-10-17 12:43:25 | agent | coder-agent | ✅ complete · T1 · 80.0s · $0.4552 |
| 2026-10-17 12:43:51 | gate | gate-build | ✅ [T1] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 12:43:51 | boss | authorize coder-agent:T2 | ✅ gate-2-architect passed |
| 2026-10-17 12:46:43 | agent | coder-agent | ✅ complete · T2 · 170.7s · $0.4902 |
| 2026-10-17 12:47:05 | gate | gate-build | ✅ [T2] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 12:47:05 | boss | authorize coder-agent:T3 | ✅ gate-2-architect passed |
| 2026-10-17 12:47:52 | agent | coder-agent | ✅ scope-request · T3 · 46.6s · $0.4237 |
| 2026-10-17 12:47:52 | gate | gate-question | ❌ REJECTED: yes: jsonschema.go granted for this run, only for the one-line switch of the meta-schema check to the no-annotation evaluation. If the meta-schema check counts annotations against the annotation-count limit … |
| 2026-10-17 12:48:24 | boss | authorize coder-agent:T3 | ✅ gate-2-architect passed |
| 2026-10-17 12:49:14 | agent | coder-agent | ✅ complete · T3 · 49.6s · $0.4092 |
| 2026-10-17 12:50:52 | gate | gate-build | ✅ [T3] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 12:50:52 | boss | authorize coder-agent:T4 | ✅ gate-2-architect passed |
| 2026-10-17 12:51:43 | agent | coder-agent | ✅ complete · T4 · 50.0s · $0.3451 |
| 2026-10-17 12:52:05 | gate | gate-build | ✅ [T4] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 12:52:05 | boss | authorize coder-agent:T5 | ✅ gate-2-architect passed |
| 2026-10-17 12:52:54 | agent | coder-agent | ✅ complete · T5 · 49.1s · $0.4460 |
| 2026-10-17 12:53:15 | gate | gate-build | ✅ [T5] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 12:53:16 | boss | authorize coder-agent:T6 | ✅ gate-2-architect passed |
| 2026-10-17 12:55:44 | agent | coder-agent | ✅ complete · T6 · 148.7s · $0.4024 |
| 2026-10-17 12:56:05 | gate | gate-build | ✅ [T6] go_test:pass, go_race:pass |
| 2026-10-17 12:56:05 | boss | authorize tester-agent | ✅ gate-build passed with all 6 task(s) implemented |
| 2026-10-17 12:57:22 | agent | tester-agent | ✅ warn · 76.3s · $0.7723 |
| 2026-10-17 12:57:22 | gate | gate-test | ✅ QA/security/performance passed (with warnings). AC covered: 8; ⚠️ 1 AC unassessed by tester: ['The shared built-in result and every compiled Schema are immutable after they are built and safe for concurrent use. A new… |
| 2026-10-17 12:57:22 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-17 12:58:10 | agent | release-agent | ✅ warn · 47.5s · $0.0785 |
| 2026-10-17 12:58:10 | gate | gate-release | ⏸️ awaiting human sign-off |
| 2026-10-17 12:58:10 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
