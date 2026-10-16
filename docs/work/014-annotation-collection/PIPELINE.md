# Pipeline record — #14: Annotation collection: metadata, format, content and unknown keywords, evaluated properties and items

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #144 · **Status:** waiting_human · **Stage:** gate-release-human
- **Started:** 2026-10-16 21:28:36 UTC · **Stopped:** 2026-10-16 21:53:00 UTC
- **Story spend:** ~$9.59 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-release: `factory approve 144` to accept, or `factory reject 144 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-release**:
  > 🚀 RELEASE SIGN-OFF (Checkpoint 3) — READY: every evidence bar is met.
  >   ◦ Release-agent note: dbg_test.go is an empty stray file with only a package declaration; it should be deleted. Tester found this as F5 (low severity, persistent).
  >   ◦ Release-agent note: Report-bytes error in run.go passes through but cannot be tested from CLI until story 193 adds the -max-report-bytes flag. Feature is covered at library level in reportbytes_test.go.
  >   ◦ Release-agent note: Breaking change: passing 'basic' reports now include annotation units (longer output). Operator explicitly approved this as matching the 2020-12 standard shape. Scripts comparing before/after output will see a difference.
  >   ◦ Release-agent note: Tester found no security issues. Build and go test -race pass. Every acceptance criterion (1-11) is covered by spec-derived tests.
  >   ◦ Security note (low): No insecure defaults found. Budget and RenderReport treat a cap of 0 or less as DefaultMaxReportBytes, never as unlimited. A negative MaxAnnotations or MaxReportBytes is a compile error. The CLI rejects a -max-annotations value of 0 or less, or one it cannot parse, with exit 2. run.go passes only the *LimitError text through and keeps 'cannot render result' for other render errors. Annotation values are schema values held by reference. Tests check that distinctive string, number and key values from the document never appear in output.
  >   → Approve to release, or reject with what must change — it goes back to the coder, then the tester.

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-16 21:28:36 | run | run started | • |
| 2026-10-16 21:28:37 | agent | spec-agent | ✅ pass |
| 2026-10-16 21:28:37 | gate | gate-1-spec | ✅ Story 'Annotation collection: metadata, format, content and unknown keywords, evaluated properties and items' has 11 AC and 6 tasks |
| 2026-10-16 21:28:37 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-16 21:28:37 | agent | architect-agent | ✅ pass |
| 2026-10-16 21:28:38 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security) |
| 2026-10-16 21:29:11 | agent | boundary-agent | ✅ warn · 33.6s · $0.2020 |
| 2026-10-16 21:29:11 | gate | gate-2-architect | ❌ REJECTED: Keep the design and settle the boundary points: (1) MaxAnnotations follows the existing limit style exactly like MaxOutputUnits: unset/0 in the library = DefaultMaxAnnotations, negative = compile error, expl… |
| 2026-10-16 21:29:38 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-16 21:30:21 | agent | architect-agent | ✅ pass · 42.5s · $0.5745 |
| 2026-10-16 21:30:21 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security) |
| 2026-10-16 21:30:40 | agent | boundary-agent | ✅ warn · 18.5s · $0.2005 |
| 2026-10-16 21:30:40 | gate | gate-2-architect | ❌ REJECTED: Keep this design and add one bound: a cap on the total bytes of a rendered report, DefaultMaxReportBytes = 256 MiB, in the existing limit style (library option, 0 = default, negative = error; CLI flag left t… |
| 2026-10-16 21:32:12 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-16 21:34:10 | agent | architect-agent | ✅ pass · 117.7s · $0.8801 |
| 2026-10-16 21:34:10 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security) |
| 2026-10-16 21:34:27 | agent | boundary-agent | ✅ pass · 16.4s · $0.2078 |
| 2026-10-16 21:34:27 | gate | gate-2-architect | ✅ APPROVED: Approved by human reviewer |
| 2026-10-16 21:34:40 | boss | authorize coder-agent:T1 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 21:36:50 | agent | coder-agent | ✅ complete · T1 · 129.2s · $0.5768 |
| 2026-10-16 21:36:57 | gate | gate-build | ✅ [T1] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 21:36:57 | boss | authorize coder-agent:T2 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 21:37:44 | agent | coder-agent | ✅ complete · T2 · 46.9s · $0.4180 |
| 2026-10-16 21:37:59 | gate | gate-build | ✅ [T2] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 21:37:59 | boss | authorize coder-agent:T3 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 21:38:57 | agent | coder-agent | ✅ complete · T3 · 58.1s · $0.4537 |
| 2026-10-16 21:39:11 | gate | gate-build | ✅ [T3] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 21:39:12 | boss | authorize coder-agent:T4 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 21:40:49 | agent | coder-agent | ✅ complete · T4 · 97.0s · $0.6153 |
| 2026-10-16 21:40:50 | gate | gate-build | ✅ [T4] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 21:40:50 | boss | authorize coder-agent:T5 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 21:42:09 | agent | coder-agent | ✅ complete · T5 · 78.0s · $0.5266 |
| 2026-10-16 21:42:16 | gate | gate-build | ✅ [T5] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 21:42:16 | boss | authorize coder-agent:T6 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 21:43:46 | agent | coder-agent | ✅ scope-request · T6 · 89.7s · $0.6040 |
| 2026-10-16 21:43:46 | gate | gate-question | ❌ REJECTED: yes, grant internal/compile/refs.go for this run (unknown keywords in $ref-compiled subtrees must be collected as annotations like in compile.go). |
| 2026-10-16 21:44:26 | boss | authorize coder-agent:T6 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 21:45:59 | agent | coder-agent | ✅ complete · T6 · 92.0s · $0.5811 |
| 2026-10-16 21:46:02 | gate | gate-build | ✅ [T6] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 21:46:02 | boss | authorize tester-agent | ✅ gate-build passed with all 6 task(s) implemented |
| 2026-10-16 21:47:33 | agent | tester-agent | ❌ fail · 90.1s · $0.9698 |
| 2026-10-16 21:47:33 | gate | gate-test | ❌ Tester overall verdict is 'fail'; QA failed; missing coverage: ['AC10: annotation values are not rendered correctly. Schema numbers are *jsonvalue.Number, which has only unexported fields and no MarshalJSON. So intern… |
| 2026-10-16 21:47:33 | boss | authorize coder-agent:remediation | ✅ gate-test failed with 11 finding(s) |
| 2026-10-16 21:51:10 | agent | coder-agent | ✅ complete · remediation · 217.0s · $1.8115 |
| 2026-10-16 21:51:12 | gate | gate-build | ✅ [remediation] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 21:51:12 | boss | authorize tester-agent | ✅ gate-build passed with all 6 task(s) implemented |
| 2026-10-16 21:51:48 | agent | tester-agent | ✅ warn · 35.8s · $0.8371 |
| 2026-10-16 21:51:48 | gate | gate-test | ✅ QA/security/performance passed (with warnings). AC covered: 11; findings: 0 new, 1 persistent, 6 resolved |
| 2026-10-16 21:51:48 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-16 21:52:59 | agent | release-agent | ✅ warn · 70.4s · $0.1285 |
| 2026-10-16 21:53:00 | gate | gate-release | ⏸️ awaiting human sign-off |
| 2026-10-16 21:53:00 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
