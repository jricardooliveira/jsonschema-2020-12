# Pipeline record — #23: Record evaluated properties and items only when something can use the record

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #159 · **Status:** waiting_human · **Stage:** gate-release-human
- **Started:** 2026-10-17 12:09:09 UTC · **Stopped:** 2026-10-17 12:34:04 UTC
- **Story spend:** ~$6.33 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-release: `factory approve 159` to accept, or `factory reject 159 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-release**:
  > 🚀 RELEASE SIGN-OFF (Checkpoint 3) — NOT READY: 1 gap(s) in the evidence:
  >   • Existing tests were changed or deleted by this story — check that they were not weakened to make a failing check pass: internal/eval/location_test.go (deleted or renamed: TestLazyLocationEqualsEager, TestLazyLocationParallelRender; assertions 3 → 0)
  >   ◦ Release-agent note: The skip switch is correctly fail-safe: its zero value is false (recording on), so only EvaluateValidate sets it. Tree and lazy evaluation paths always record, preserving all outputs.
  >   ◦ Release-agent note: The load-time analysis recurses depth-first over the already-bounded schema and allocates nothing (no heap stack), so CompileOrder allocs/op stays at the baseline.
  >   ◦ Release-agent note: The old location.go tests for dead nodeLoc/kwLoc types were replaced by lazyloc_test.go, which tests the live lazyLoc implementation through the real evaluator. This is a legitimate test upgrade per AC6 (remove unused types and test lazyLoc); the new test covers all required cases.
  >   ◦ Release-agent note: Benchmark improvements are primarily in memory (B/op and allocs/op), which was the goal. Speed improvements (ns/op) are within the ±5% noise band and are recorded for reference only.
  >   ◦ Security note (none): No insecure default. The per-call switch is fail-safe: its zero value means recording stays on, and only EvaluateValidate sets it. NeedsEvaluated answers 'needed' when the registry cannot be listed, and checks every registry resource whenever a reference leaves the document. Step, depth, ref-chain, output-unit, report-byte and annotation accounting stay outside the guarded helpers, and both limit tests show the boundaries are unchanged. scan() now recurses, but only over a schema value that was already decoded under MaxDepth (or an embedded meta-schema), and compile and the meta-schema check already walk it to the same depth. It adds no new stack-exhaustion risk. No instance value reaches any output.
  >   → Approve to release accepting the gaps above as a known risk, or reject with what must change — it goes back to the coder, then the tester.

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-17 12:09:09 | run | run started | • |
| 2026-10-17 12:09:10 | agent | spec-agent | ✅ pass |
| 2026-10-17 12:09:10 | gate | gate-1-spec | ✅ Story 'Record evaluated properties and items only when something can use the record' has 8 AC and 5 tasks |
| 2026-10-17 12:09:10 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-17 12:09:10 | agent | architect-agent | ✅ pass |
| 2026-10-17 12:09:10 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it touches sensitive areas (security) |
| 2026-10-17 12:09:28 | agent | boundary-agent | ✅ pass · 17.4s · $0.1902 |
| 2026-10-17 12:09:28 | gate | gate-2-architect | ✅ Architecture approved. 14 modules, 4 risks flagged |
| 2026-10-17 12:09:28 | boss | authorize coder-agent:T1 | ✅ gate-2-architect passed |
| 2026-10-17 12:10:55 | agent | coder-agent | ✅ complete · T1 · 86.2s · $0.5081 |
| 2026-10-17 12:12:31 | gate | gate-build | ✅ [T1] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 12:12:31 | boss | authorize coder-agent:T4 | ✅ gate-2-architect passed |
| 2026-10-17 12:13:17 | agent | coder-agent | ✅ complete · T4 · 45.5s · $0.4061 |
| 2026-10-17 12:14:54 | gate | gate-build | ✅ [T4] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 12:14:54 | boss | authorize coder-agent:T2 | ✅ gate-2-architect passed |
| 2026-10-17 12:16:11 | agent | coder-agent | ✅ complete · T2 · 76.1s · $0.5072 |
| 2026-10-17 12:17:47 | gate | gate-build | ✅ [T2] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 12:17:47 | boss | authorize coder-agent:T3 | ✅ gate-2-architect passed |
| 2026-10-17 12:19:29 | agent | coder-agent | ✅ complete · T3 · 101.4s · $0.5581 |
| 2026-10-17 12:19:54 | gate | gate-build | ✅ [T3] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 12:19:54 | boss | authorize coder-agent:T5 | ✅ gate-2-architect passed |
| 2026-10-17 12:24:41 | agent | coder-agent | ✅ complete · T5 · 286.9s · $0.5098 |
| 2026-10-17 12:25:02 | gate | gate-build | ✅ [T5] go_test:pass, go_race:pass |
| 2026-10-17 12:25:03 | boss | authorize tester-agent | ✅ gate-build passed with all 5 task(s) implemented |
| 2026-10-17 12:26:21 | agent | tester-agent | ❌ fail · 77.9s · $0.7457 |
| 2026-10-17 12:26:21 | gate | gate-test | ❌ Tester overall verdict is 'fail'; QA failed; missing coverage: ["AC7 – 'No scenario's allocs/op or B/op rises' is not met. docs/benchmarks.md itself records CompileOrder allocs/op going from 14,072 to 14,075 and B/op … |
| 2026-10-17 12:26:21 | boss | authorize coder-agent:remediation | ✅ gate-test failed with 8 finding(s) |
| 2026-10-17 12:32:03 | agent | coder-agent | ✅ complete · remediation · 340.5s · $2.0961 |
| 2026-10-17 12:32:29 | gate | gate-build | ✅ [remediation] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 12:32:29 | boss | authorize tester-agent | ✅ gate-build passed with all 5 task(s) implemented |
| 2026-10-17 12:33:26 | agent | tester-agent | ✅ pass · 56.6s · $0.7259 |
| 2026-10-17 12:33:26 | gate | gate-test | ✅ QA/security/performance passed. AC covered: 8; findings: 0 new, 0 persistent, 8 resolved |
| 2026-10-17 12:33:26 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-17 12:34:03 | agent | release-agent | ✅ pass · 36.9s · $0.0792 |
| 2026-10-17 12:34:04 | gate | gate-release | ⏸️ awaiting human sign-off |
| 2026-10-17 12:34:04 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
