# Pipeline record — #24: Collect and sort annotations only when they can be shown, without re-splitting JSON Pointers

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #158 · **Status:** waiting_human · **Stage:** gate-release-human
- **Started:** 2026-10-17 11:34:09 UTC · **Stopped:** 2026-10-17 12:02:27 UTC
- **Story spend:** ~$4.47 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-release: `factory approve 158` to accept, or `factory reject 158 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-release**:
  > 🚀 RELEASE SIGN-OFF (Checkpoint 3) — READY: every evidence bar is met.
  >   ◦ Release-agent note: A panic inside the sync.Once guard during first annotation access would not propagate to the caller (the once prevents re-entry). This is accepted as very unlikely but limits error observability.
  >   ◦ Release-agent note: ValidateErrors benchmark shows a 3-allocation rise (510,077 to 510,080) attributed to noise, but sits just outside the before range (510,076–510,079). Monitor to confirm this does not grow further in follow-up work.
  >   ◦ Release-agent note: Segment-based sorting precomputes all pointer tokens upfront, which saves re-splitting on comparison but consumes more temporary memory during the sort. Net benefit is positive on large annotation sets.
  >   ◦ Release-agent note: eval.go and annotations.go are also touched by stories 199 and 200; later rebases may cause merge conflicts.
  >   ◦ Security note (low): No secrets, network access or file reads are added. Records and built annotations hold only locations, the keyword name and schema-sourced values; the renderers and errors are unchanged, so no instance values leak. Discard mode still counts every Annotate call against the cap before it returns, so 'not' does not open a way around the annotation-count limit.; Hardening (low): Lazy.Sorted runs sortInPlace inside sync.Once. If the sort ever panicked, the assignment would not happen but the Once would count as done, so every later reader would silently get an empty annotation list. No half-built list is published, but the failure is swallowed rather than repeated. sortInPlace has no realistic panic path, so this is defence-in-depth only.
  >   → Approve to release, or reject with what must change — it goes back to the coder, then the tester.

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-17 11:34:09 | run | run started | • |
| 2026-10-17 11:34:09 | agent | spec-agent | ✅ pass |
| 2026-10-17 11:34:10 | gate | gate-1-spec | ✅ Story 'Collect and sort annotations only when they can be shown, without re-splitting JSON Pointers' has 8 AC and 6 tasks |
| 2026-10-17 11:34:10 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-17 11:34:10 | agent | architect-agent | ✅ pass |
| 2026-10-17 11:34:10 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it touches sensitive areas (security) |
| 2026-10-17 11:34:24 | agent | boundary-agent | ✅ pass · 13.3s · $0.1780 |
| 2026-10-17 11:34:24 | gate | gate-2-architect | ✅ Architecture approved. 13 modules, 4 risks flagged |
| 2026-10-17 11:34:24 | boss | authorize coder-agent:T1 | ✅ gate-2-architect passed |
| 2026-10-17 11:35:00 | agent | coder-agent | ✅ complete · T1 · 34.9s · $0.3214 |
| 2026-10-17 11:36:40 | gate | gate-build | ✅ [T1] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 11:36:40 | boss | authorize coder-agent:T2 | ✅ gate-2-architect passed |
| 2026-10-17 11:37:41 | agent | coder-agent | ✅ scope-request · T2 · 60.3s · $0.4575 |
| 2026-10-17 11:37:41 | gate | gate-question | ❌ REJECTED: yes: output.go granted for this run, only the one-line change so Output renders from the same sorted annotations as the Result (no other change; output stays byte-identical). |
| 2026-10-17 11:38:12 | boss | authorize coder-agent:T2 | ✅ gate-2-architect passed |
| 2026-10-17 11:39:30 | agent | coder-agent | ✅ complete · T2 · 77.5s · $0.4197 |
| 2026-10-17 11:41:03 | gate | gate-build | ✅ [T2] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 11:41:04 | boss | authorize coder-agent:T3 | ✅ gate-2-architect passed |
| 2026-10-17 11:41:42 | agent | coder-agent | ✅ scope-request · T3 · 37.8s · $0.3682 |
| 2026-10-17 11:41:42 | gate | gate-question | ❌ REJECTED: yes: internal/keywords/annotate.go granted for this run, so a discarded annotation is still counted against the annotation-count limit at exactly the same point (count without appending); no other behaviour … |
| 2026-10-17 11:42:05 | boss | authorize coder-agent:T3 | ✅ gate-2-architect passed |
| 2026-10-17 11:43:07 | agent | coder-agent | ✅ complete · T3 · 62.2s · $0.4554 |
| 2026-10-17 11:44:47 | gate | gate-build | ✅ [T3] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 11:44:47 | boss | authorize coder-agent:T4 | ✅ gate-2-architect passed |
| 2026-10-17 11:45:37 | agent | coder-agent | ✅ complete · T4 · 49.4s · $0.3986 |
| 2026-10-17 11:46:02 | gate | gate-build | ✅ [T4] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 11:46:02 | boss | authorize coder-agent:T5 | ✅ gate-2-architect passed |
| 2026-10-17 11:46:34 | agent | coder-agent | ✅ complete · T5 · 31.1s · $0.3414 |
| 2026-10-17 11:46:57 | gate | gate-build | ✅ [T5] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 11:46:57 | boss | authorize coder-agent:T6 | ✅ gate-2-architect passed |
| 2026-10-17 12:00:00 | agent | coder-agent | ✅ complete · T6 · 782.1s · $0.7749 |
| 2026-10-17 12:00:31 | gate | gate-build | ✅ [T6] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 12:00:31 | boss | authorize tester-agent | ✅ gate-build passed with all 6 task(s) implemented |
| 2026-10-17 12:01:35 | agent | tester-agent | ✅ warn · 63.9s · $0.6666 |
| 2026-10-17 12:01:35 | gate | gate-test | ✅ QA/security/performance passed (with warnings). AC covered: 8; findings: 3 new, 0 persistent, 0 resolved |
| 2026-10-17 12:01:35 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-17 12:02:26 | agent | release-agent | ✅ warn · 50.6s · $0.0836 |
| 2026-10-17 12:02:27 | gate | gate-release | ⏸️ awaiting human sign-off |
| 2026-10-17 12:02:27 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
