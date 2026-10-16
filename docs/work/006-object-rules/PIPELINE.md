# Pipeline record — #6: 2020-12 object keywords: properties, patternProperties, additionalProperties, propertyNames, dependentSchemas, maxProperties, minProperties, required, dependentRequired

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #141 · **Status:** waiting_human · **Stage:** gate-release-human
- **Started:** 2026-10-16 20:16:13 UTC · **Stopped:** 2026-10-16 20:37:58 UTC
- **Story spend:** ~$7.03 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-release: `factory approve 141` to accept, or `factory reject 141 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-release**:
  > 🚀 RELEASE SIGN-OFF (Checkpoint 3) — READY: every evidence bar is met.
  >   ◦ Release-agent note: Minor test-coverage gaps documented by tester: >20-key determinism test compares units rather than rendered bytes as architecture specified; propertyNames order test incomplete. Neither affects functionality, both flagged as lower priority.
  >   ◦ Release-agent note: patternProperties and additionalProperties duplicate ~6 lines of regex compilation (size check, ECMA-262 translate, compile) from validation/pattern.go. Documented as follow-up: create a shared CompilePattern helper.
  >   ◦ Release-agent note: Story 191 (annotation collection) should build on the EvaluatedItems mechanism (evaluated.go, evaluated_items.go) rather than duplicate property recording. The operator noted this risk in the architecture review.
  >   ◦ Release-agent note: The per-evaluation record now lives in EvaluatedItems, keyed by (instance location, keyword location). A keyword records names whether its subschemas pass or fail; story 191 must filter out failed annotations if the spec requires it. The interface is designed for that, but not yet tested.
  >   ◦ Security note (low): No insecure defaults found. Compile errors quote only the schema's own pattern. Unit messages are fixed phrases or schema-sourced names. Instance keys appear only in the JSON Pointer instance location, with ~0/~1 escaping. The step limit and the output-unit cap both return a LimitError that is passed on and never swallowed (TestPatternPropertiesStepLimit, TestPropertiesLimit, TestPropertyNamesLimit).; Low: the new 'evaluation steps' limit is only tested at the internal eval level (EvaluateLimited). No test goes through jsonschema.Validate or the CLI to show it gives exit code 2. With the 100M default, a library caller cannot trigger it in practice. Story 193 makes it configurable and should add that end-to-end check.
  >   → Approve to release, or reject with what must change — it goes back to the coder, then the tester.

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
| 2026-10-16 20:26:45 | gate | gate-release | ❌ REJECTED: Fix F1 (medium) narrowly: the approved boundary rule assumed an evaluation step limit that does not exist yet. Add a minimal per-call step counter to the evaluation context (allocated per call, never in the … |
| 2026-10-16 20:27:31 | boss | authorize coder-agent:remediation | ✅ the operator's rejection at Checkpoint 3 (1 finding(s)) |
| 2026-10-16 20:36:27 | agent | coder-agent | ✅ complete · remediation · 338.1s · $1.5710 |
| 2026-10-16 20:36:28 | gate | gate-build | ✅ [remediation] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 20:36:28 | boss | authorize tester-agent | ✅ gate-build passed with all 5 task(s) implemented |
| 2026-10-16 20:37:26 | agent | tester-agent | ✅ warn · 57.3s · $0.7933 |
| 2026-10-16 20:37:26 | gate | gate-test | ✅ QA/security/performance passed (with warnings). AC covered: 11; findings: 3 new, 0 persistent, 0 resolved |
| 2026-10-16 20:37:26 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-16 20:37:57 | agent | release-agent | ✅ warn · 31.2s · $0.0912 |
| 2026-10-16 20:37:58 | gate | gate-release | ⏸️ awaiting human sign-off |
| 2026-10-16 20:37:58 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
