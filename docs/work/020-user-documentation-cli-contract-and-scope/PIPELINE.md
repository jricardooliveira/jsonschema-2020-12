# Pipeline record — #20: README and docs/ that describe the finished validator, proven by tests

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #155 · **Status:** waiting_human · **Stage:** gate-release-human
- **Started:** 2026-10-17 05:21:08 UTC · **Stopped:** 2026-10-17 05:33:12 UTC
- **Story spend:** ~$6.31 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-release: `factory approve 155` to accept, or `factory reject 155 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-release**:
  > 🚀 RELEASE SIGN-OFF (Checkpoint 3) — READY: every evidence bar is met.
  >   ◦ Release-agent note: Limit defaults are documented as read from the code, but there is no CI test that parses the docs and compares them to the actual code values. The docs are currently accurate (verified by manual inspection), but a future change to a default value (e.g. story 193's flags or a later story) could drift undetected. This is a low-severity issue because: (1) the flag test catches removed or added flags, (2) defaults are less likely to change than flags, and (3) the concern is noted for future stories. A test that extracts defaults from docs/limits.md and compares them to the actual option and constant values could be added in a follow-up story.
  >   ◦ Release-agent note: The diff shows docs/limits.md was significantly rewritten (line count ~50 to ~60), but the exact state before this story is not visible. The story depends on stories 178–196, so all limits should be present. Verify that every constant in options.go, numberlimits.go, cmd/jsonschema/limitflags.go and cmd/jsonschema/refflags.go appears in the table.
  >   ◦ Release-agent note: The docs_examples_test.go parses CLI examples as ```sh and ```json blocks in markdown. If a future page uses a code fence with trailing spaces or unusual formatting, the parser might fail silently (the test would see fewer examples than written). The test checks `if n == 0 { t.Fatal(...) }`, so it will error if no examples are found, but a single formatting error in one block is not caught. Consider a lint pass over the markdown if docs grow much larger.
  >   → Approve to release, or reject with what must change — it goes back to the coder, then the tester.

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-17 05:21:08 | run | run started | • |
| 2026-10-17 05:21:09 | agent | spec-agent | ✅ pass |
| 2026-10-17 05:21:09 | gate | gate-1-spec | ✅ APPROVED: Approved by human reviewer |
| 2026-10-17 05:21:47 | boss | authorize architect-agent | ✅ gate-1-spec, approved by the operator at Checkpoint 1 |
| 2026-10-17 05:21:47 | agent | architect-agent | ✅ pass |
| 2026-10-17 05:21:47 | gate | gate-2-architect | ✅ Architecture approved. 15 modules, 3 risks flagged |
| 2026-10-17 05:21:47 | boss | authorize coder-agent:T1 | ✅ gate-2-architect passed |
| 2026-10-17 05:22:04 | agent | coder-agent | ✅ complete · T1 · 16.3s · $0.2150 |
| 2026-10-17 05:22:27 | gate | gate-build | ✅ [T1] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 05:22:28 | boss | authorize coder-agent:T2 | ✅ gate-2-architect passed |
| 2026-10-17 05:23:41 | agent | coder-agent | ✅ complete · T2 · 73.0s · $0.4276 |
| 2026-10-17 05:23:41 | gate | gate-build | ✅ [T2] go_test:pass, go_race:pass |
| 2026-10-17 05:23:41 | boss | authorize coder-agent:T3 | ✅ gate-2-architect passed |
| 2026-10-17 05:25:11 | agent | coder-agent | ✅ complete · T3 · 89.6s · $0.5175 |
| 2026-10-17 05:25:11 | gate | gate-build | ❌ [T3] SCOPE: writes outside the declared files were refused: ['jsbin'] (allowed: ['docs/output-formats.md', 'docs/scope.md', 'docs/deviations.md', 'cmd/jsonschema/run.go', 'README.md', 'docs/cli.md', 'docs/limits.md', … |
| 2026-10-17 05:25:11 | boss | authorize coder-agent:T3 | ✅ gate-2-architect passed |
| 2026-10-17 05:25:54 | agent | coder-agent | ✅ complete · T3 · 42.5s · $0.7751 |
| 2026-10-17 05:25:54 | gate | gate-build | ✅ [T3] go_test:pass, go_race:pass |
| 2026-10-17 05:25:54 | boss | authorize coder-agent:T4 | ✅ gate-2-architect passed |
| 2026-10-17 05:26:19 | agent | coder-agent | ✅ complete · T4 · 24.7s · $0.3024 |
| 2026-10-17 05:26:39 | gate | gate-build | ✅ [T4] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 05:26:39 | boss | authorize coder-agent:T6 | ✅ gate-2-architect passed |
| 2026-10-17 05:27:20 | agent | coder-agent | ✅ complete · T6 · 40.3s · $0.3412 |
| 2026-10-17 05:27:42 | gate | gate-build | ✅ [T6] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 05:27:42 | boss | authorize coder-agent:T5 | ✅ gate-2-architect passed |
| 2026-10-17 05:28:07 | agent | coder-agent | ✅ complete · T5 · 24.9s · $0.2966 |
| 2026-10-17 05:28:30 | gate | gate-build | ✅ [T5] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 05:28:30 | boss | authorize tester-agent | ✅ gate-build passed with all 6 task(s) implemented |
| 2026-10-17 05:29:20 | agent | tester-agent | ❌ fail · 50.3s · $0.6904 |
| 2026-10-17 05:29:20 | gate | gate-test | ❌ Tester overall verdict is 'fail'; QA failed; missing coverage: ["AC9 (structural, visibly not met): the earlier pages docs/regex-deviations.md, docs/formats.md and docs/cli-contract.md are still on disk. Their content… |
| 2026-10-17 05:29:21 | boss | authorize coder-agent:remediation | ✅ gate-test failed with 5 finding(s) |
| 2026-10-17 05:30:18 | agent | coder-agent | ✅ scope-request · remediation · 57.0s · $0.9380 |
| 2026-10-17 05:30:18 | gate | gate-question | ❌ REJECTED: yes, grant docs/regex-deviations.md, docs/formats.md, docs/cli-contract.md and docs/references.md for this run. File deletion is not available in this session: instead of deleting a stale merged page, replac… |
| 2026-10-17 05:30:36 | boss | authorize coder-agent:remediation | ✅ gate-test failed with 3 finding(s) |
| 2026-10-17 05:31:40 | agent | coder-agent | ✅ complete · remediation · 63.9s · $1.0418 |
| 2026-10-17 05:32:02 | gate | gate-build | ✅ [remediation] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 05:32:02 | boss | authorize tester-agent | ✅ gate-build passed with all 6 task(s) implemented |
| 2026-10-17 05:32:32 | agent | tester-agent | ✅ pass · 30.0s · $0.6722 |
| 2026-10-17 05:32:32 | gate | gate-test | ✅ QA/security/performance passed. AC covered: 15; findings: 1 new, 0 persistent, 3 resolved |
| 2026-10-17 05:32:33 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-17 05:33:12 | agent | release-agent | ✅ warn · 38.9s · $0.0881 |
| 2026-10-17 05:33:12 | gate | gate-release | ⏸️ awaiting human sign-off |
| 2026-10-17 05:33:12 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
