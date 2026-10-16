# Pipeline record — #3: Error reports: flag and basic output, multiple documents

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #137 · **Status:** waiting_human · **Stage:** gate-release-human
- **Started:** 2026-10-16 19:00:55 UTC · **Stopped:** 2026-10-16 19:53:16 UTC
- **Story spend:** ~$4.11 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-release: `factory approve 137` to accept, or `factory reject 137 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-release**:
  > 🚀 RELEASE SIGN-OFF (Checkpoint 3) — READY: every evidence bar is met.
  >   ◦ Release-agent note: Breaking change in CLI stdout: Single invalid documents now produce the full 2020-12 basic output instead of the minimal result. This is an accepted breaking change (documented in the architecture), but scripts that parse old output must be updated.
  >   ◦ Release-agent note: Per-document error messages are built directly from Go's err.Error() without filtering. File system paths or other system details may appear in stderr (e.g., 'open /path/to/file: permission denied'). The tester approved this as acceptable.
  >   ◦ Release-agent note: Output-unit cap memory usage: The 10M default cap, if fully consumed, could accumulate millions of units in memory on very large schemas. This is generous for legitimate use but worth monitoring if extremely large documents are validated.
  >   ◦ Release-agent note: Future applicators (story 185) must use the same EvalContext.EvalChild() method to evaluate child schemas, or the output-unit cap could be bypassed. The regression test ensures discarded units are counted; story 185 must maintain this contract.
  >   ◦ Security note (low): F2 (low): run.go fills per-document NDJSON and stderr reasons from err.Error(). Today this is safe: Validate wraps only the content-free *SyntaxError and *LimitError with %w, and readFile/readCapped produce fixed phrases. Nothing enforces the boundary rule that only typed errors may reach the output, though. A later untyped error from Validate (for example from a future keyword's compile or eval path) would go straight to stdout and stderr. A typed allow-list, or a generic fallback reason for anything that is not a known type, would make this fail closed.
  >   → Approve to release, or reject with what must change — it goes back to the coder, then the tester.

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-16 19:00:55 | run | run started | • |
| 2026-10-16 19:00:55 | agent | spec-agent | ✅ pass |
| 2026-10-16 19:00:55 | gate | gate-1-spec | ✅ APPROVED: Approved by human reviewer |
| 2026-10-16 19:01:20 | boss | authorize architect-agent | ✅ gate-1-spec, approved by the operator at Checkpoint 1 |
| 2026-10-16 19:01:20 | agent | architect-agent | ✅ pass |
| 2026-10-16 19:01:20 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security, pii) |
| 2026-10-16 19:01:44 | agent | boundary-agent | ✅ pass · 23.4s · $0.1630 |
| 2026-10-16 19:01:44 | gate | gate-2-architect | ✅ APPROVED: Approved by human reviewer |
| 2026-10-16 19:02:03 | boss | authorize coder-agent:T1 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 19:04:02 | agent | coder-agent | ✅ complete · T1 · 119.4s · $0.4623 |
| 2026-10-16 19:04:06 | gate | gate-build | ✅ [T1] go_build:pass, go_vet:pass, go_test:pass |
| 2026-10-16 19:04:06 | boss | authorize coder-agent:T2 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 19:04:47 | agent | coder-agent | ✅ complete · T2 · 40.5s · $0.3135 |
| 2026-10-16 19:04:48 | gate | gate-build | ✅ [T2] go_build:pass, go_vet:pass, go_test:pass |
| 2026-10-16 19:04:48 | boss | authorize coder-agent:T3 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 19:05:36 | agent | coder-agent | ✅ complete · T3 · 48.4s · $0.3306 |
| 2026-10-16 19:05:39 | gate | gate-build | ✅ [T3] go_build:pass, go_vet:pass, go_test:pass |
| 2026-10-16 19:05:39 | boss | authorize coder-agent:T4 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 19:06:30 | agent | coder-agent | ✅ complete · T4 · 50.6s · $0.3528 |
| 2026-10-16 19:06:31 | gate | gate-build | ✅ [T4] go_build:pass, go_vet:pass, go_test:pass |
| 2026-10-16 19:06:32 | boss | authorize coder-agent:T5 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 19:07:07 | agent | coder-agent | ✅ complete · T5 · 35.8s · $0.3025 |
| 2026-10-16 19:07:10 | gate | gate-build | ✅ [T5] go_build:pass, go_vet:pass, go_test:pass |
| 2026-10-16 19:07:10 | boss | authorize tester-agent | ✅ gate-build passed with all 5 task(s) implemented |
| 2026-10-16 19:08:29 | agent | tester-agent | ❌ error |
| 2026-10-16 19:47:35 | boss | authorize tester-agent | ✅ gate-build passed with all 5 task(s) implemented |
| 2026-10-16 19:48:46 | agent | tester-agent | ✅ warn · 70.2s · $0.6839 |
| 2026-10-16 19:48:46 | gate | gate-test | ✅ QA/security/performance passed (with warnings). AC covered: 11; findings: 3 new, 0 persistent, 1 resolved |
| 2026-10-16 19:48:46 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-16 19:49:33 | agent | release-agent | ✅ warn · 46.6s · $0.0824 |
| 2026-10-16 19:49:34 | gate | gate-release | ❌ REJECTED: One fix only, change nothing else: internal/eval/units.go sets DefaultMaxOutputUnits = 100000, but the design this relaunch keeps (run #136) fixed it at 10_000_000 so that large but legitimate invalid docume… |
| 2026-10-16 19:51:09 | boss | authorize coder-agent:remediation | ✅ the operator's rejection at Checkpoint 3 (1 finding(s)) |
| 2026-10-16 19:51:29 | agent | coder-agent | ✅ complete · remediation · 19.5s · $0.7340 |
| 2026-10-16 19:51:32 | gate | gate-build | ✅ [remediation] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 19:51:33 | boss | authorize tester-agent | ✅ gate-build passed with all 5 task(s) implemented |
| 2026-10-16 19:52:18 | agent | tester-agent | ✅ pass · 45.2s · $0.5930 |
| 2026-10-16 19:52:18 | gate | gate-test | ✅ QA/security/performance passed. AC covered: 11; findings: 3 new, 0 persistent, 0 resolved |
| 2026-10-16 19:52:18 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-16 19:53:16 | agent | release-agent | ✅ pass · 57.4s · $0.0894 |
| 2026-10-16 19:53:16 | gate | gate-release | ⏸️ awaiting human sign-off |
| 2026-10-16 19:53:16 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
