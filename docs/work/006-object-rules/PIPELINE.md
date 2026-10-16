# Pipeline record — #6: 2020-12 object keywords: properties, patternProperties, additionalProperties, propertyNames, dependentSchemas, maxProperties, minProperties, required, dependentRequired

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #141 · **Status:** waiting_human · **Stage:** gate-release-human
- **Started:** 2026-10-16 20:16:13 UTC · **Stopped:** 2026-10-16 20:26:45 UTC
- **Story spend:** ~$4.57 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-release: `factory approve 141` to accept, or `factory reject 141 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-release**:
  > 🚀 RELEASE SIGN-OFF (Checkpoint 3) — NOT READY: 1 gap(s) in the evidence:
  >   • Security finding (medium) not resolved: Boundary rule not met: every member visited and every pattern match attempt was supposed to count toward an evaluation step limit. No such limit exists in the engine (EvalContext has only the output-unit counter), and the code does not add one. Only failures are bounded, by the output-unit cap, which TestPropertiesLimit and TestPropertyNamesLimit check. A hostile schema with many large patternProperties (each up to DefaultMaxPatternSize) against an object with many members that all pass does members×patterns RE2 matches and never gets a LimitError. RE2 is linear and input size is capped, so this is bounded CPU cost, not a hang (medium)
  >   ◦ Release-agent note: Unbounded evaluation on large objects: the tester flagged that object keyword processing (especially patternProperties with many patterns) does not count toward the evaluation step limit. Running a schema with 50 patternProperties against an object with 1000 keys causes 50,000 regex matches, unbounded by evaluation limits. RE2 is linear so it does not hang, and input size is capped by a separate limit, so this is a resource-contention issue rather than a crash risk. However, a service accepting untrusted schemas could be DOS'd by a schema with many patterns. Consider a follow-up story to add step counting to object keywords.
  >   ◦ Release-agent note: The per-evaluation property record (evaluated_items.go) is outside the original story scope but required to avoid leaking annotations on failed subschemas. The change is small (three fields in itemEntry), and the file is not crowded, but verify in code review that the addition is minimal.
  >   ◦ Release-agent note: Story 191 (annotation collection and filtering) will define its own mechanism. It should build on the EvaluatedItems log (evaluated.go and evaluated_items.go) rather than duplicate state, to avoid data loss when subschemas fail.
  >   → Approve to release accepting the gaps above as a known risk, or reject with what must change — it goes back to the coder, then the tester.

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-16 20:16:13 | run | run started | • |
| 2026-10-16 20:16:13 | agent | spec-agent | ✅ pass |
| 2026-10-16 20:16:13 | gate | gate-1-spec | ✅ Story '2020-12 object keywords: properties, patternProperties, additionalProperties, propertyNames, dependentSchemas, maxProperties, minProperties, required, dependentRequired' has 11 AC and 5 tasks |
| 2026-10-16 20:16:13 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-16 20:16:13 | agent | architect-agent | ✅ pass |
| 2026-10-16 20:16:14 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it touches sensitive areas (pii) |
| 2026-10-16 20:16:32 | agent | boundary-agent | ✅ warn · 18.6s · $0.1741 |
| 2026-10-16 20:16:32 | gate | gate-2-architect | ❌ REJECTED: Keep the design and settle the ordering point without touching the decoder: iterate object members in deterministic sorted key order (byte-wise UTF-8 order of the member names), not document order (the opera… |
| 2026-10-16 20:17:31 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-16 20:18:47 | agent | architect-agent | ✅ pass · 76.5s · $0.6877 |
| 2026-10-16 20:18:48 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it touches sensitive areas (pii) |
| 2026-10-16 20:19:10 | agent | boundary-agent | ✅ warn · 22.8s · $0.1972 |
| 2026-10-16 20:19:11 | gate | gate-2-architect | ❌ REJECTED: Accept the design's order as THE rule (operator decision, supersedes my previous wording): keywords are evaluated and reported in sorted keyword-name order; within one keyword, member failures come in sorted… |
| 2026-10-16 20:19:31 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-16 20:20:16 | agent | architect-agent | ✅ pass · 44.8s · $0.4578 |
| 2026-10-16 20:20:16 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it touches sensitive areas (pii) |
| 2026-10-16 20:20:32 | agent | boundary-agent | ✅ pass · 15.4s · $0.1899 |
| 2026-10-16 20:20:32 | gate | gate-2-architect | ✅ Architecture approved. 30 modules, 5 risks flagged |
| 2026-10-16 20:20:32 | boss | authorize coder-agent:T1 | ✅ gate-2-architect passed |
| 2026-10-16 20:21:07 | agent | coder-agent | ✅ complete · T1 · 34.9s · $0.3053 |
| 2026-10-16 20:21:13 | gate | gate-build | ✅ [T1] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 20:21:13 | boss | authorize coder-agent:T3 | ✅ gate-2-architect passed |
| 2026-10-16 20:22:15 | agent | coder-agent | ✅ complete · T3 · 62.1s · $0.4151 |
| 2026-10-16 20:22:20 | gate | gate-build | ✅ [T3] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 20:22:21 | boss | authorize coder-agent:T4 | ✅ gate-2-architect passed |
| 2026-10-16 20:23:34 | agent | coder-agent | ✅ complete · T4 · 73.3s · $0.4533 |
| 2026-10-16 20:23:39 | gate | gate-build | ✅ [T4] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 20:23:40 | boss | authorize coder-agent:T2 | ✅ gate-2-architect passed |
| 2026-10-16 20:24:58 | agent | coder-agent | ✅ complete · T2 · 78.2s · $0.4823 |
| 2026-10-16 20:24:59 | gate | gate-build | ✅ [T2] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 20:24:59 | boss | authorize coder-agent:T5 | ✅ gate-2-architect passed |
| 2026-10-16 20:25:22 | agent | coder-agent | ✅ complete · T5 · 22.3s · $0.3206 |
| 2026-10-16 20:25:25 | gate | gate-build | ✅ [T5] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 20:25:25 | boss | authorize tester-agent | ✅ gate-build passed with all 5 task(s) implemented |
| 2026-10-16 20:26:16 | agent | tester-agent | ✅ warn · 51.0s · $0.8032 |
| 2026-10-16 20:26:17 | gate | gate-test | ✅ QA/security/performance passed (with warnings). AC covered: 11; findings: 3 new, 0 persistent, 0 resolved |
| 2026-10-16 20:26:17 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-16 20:26:45 | agent | release-agent | ✅ warn · 28.1s · $0.0851 |
| 2026-10-16 20:26:45 | gate | gate-release | ⏸️ awaiting human sign-off |
| 2026-10-16 20:26:45 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
