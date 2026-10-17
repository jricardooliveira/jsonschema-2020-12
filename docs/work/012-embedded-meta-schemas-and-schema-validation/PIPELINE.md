# Pipeline record — #12: Embedded 2020-12 meta-schemas and meta-schema validation of every loaded schema

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #148 · **Status:** waiting_human · **Stage:** gate-release-human
- **Started:** 2026-10-17 00:11:06 UTC · **Stopped:** 2026-10-17 00:37:32 UTC
- **Story spend:** ~$9.61 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-release: `factory approve 148` to accept, or `factory reject 148 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-release**:
  > 🚀 RELEASE SIGN-OFF (Checkpoint 3) — NOT READY: 1 gap(s) in the evidence:
  >   • 1 acceptance criteria are not proven by any test the tester could see (§5.1) — verify them yourself or send the story back: ["The meta-schema check never asserts 'format', whether or not format-assertion mode is on, so enabling format assertion never makes a previously accepted schema fail."]
  >   ◦ Release-agent note: Acceptance criterion 7 (meta-schema check never asserts format, whether format-assertion mode is on or off) is not proven by any test, because format-assertion mode does not exist. The implementation intends to force format off (architecture notes say so), and the code calls eval.EvaluateFull without a visible format parameter, but this needs verification in story 195.
  >   ◦ Release-agent note: Two removed test cases (negative minItems, fractional minItems) are now schema errors instead of validation results. This is correct behaviour — they belong in metaschema_test.go. However, the spec_test.go runner now skips the entire metaschema subdirectory to avoid running them twice; ensure this does not hide any other test case changes.
  >   ◦ Release-agent note: Built-in meta-schemas and the root meta-schema are recompiled on every Compile call (see metaCheck in jsonschema.go). This is correct but not optimal for repeated compilations with the same registry. A future optimization could cache compiled meta-schemas.
  >   ◦ Security note (low): The meta-check is fail-closed: MetaChecks returns any error from ResolveMetaSchema, CompileWithRegistry or the evaluator, and metaCheck wraps non-ISE errors as 'schema: ...'. Builtins() returns its load error rather than skipping. The chain is bounded by maxMetaChain=32 with a seen set. Embedded meta-schemas are read only through go:embed, and custom ones only from the supplied index. The unsupported and chain errors are constant strings that never echo the declared '$schema'. InvalidSchemaError holds only the source label, the instance pointer and the keyword name.
  >   → Approve to release accepting the gaps above as a known risk, or reject with what must change — it goes back to the coder, then the tester.

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-17 00:11:06 | run | run started | • |
| 2026-10-17 00:11:06 | agent | spec-agent | ✅ pass |
| 2026-10-17 00:11:06 | gate | gate-1-spec | ✅ Story 'Embedded 2020-12 meta-schemas and meta-schema validation of every loaded schema' has 11 AC and 5 tasks |
| 2026-10-17 00:11:06 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-17 00:11:06 | agent | architect-agent | ✅ pass |
| 2026-10-17 00:11:07 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security) |
| 2026-10-17 00:11:27 | agent | boundary-agent | ✅ pass · 19.9s · $0.1785 |
| 2026-10-17 00:11:27 | gate | gate-2-architect | ✅ APPROVED: Approved by human reviewer |
| 2026-10-17 00:11:44 | boss | authorize coder-agent:T1 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-17 00:13:14 | agent | coder-agent | ✅ complete · T1 · 89.5s · $0.4589 |
| 2026-10-17 00:13:33 | gate | gate-build | ✅ [T1] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 00:13:33 | boss | authorize coder-agent:T2 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-17 00:14:25 | agent | coder-agent | ✅ complete · T2 · 51.3s · $0.3543 |
| 2026-10-17 00:14:45 | gate | gate-build | ✅ [T2] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 00:14:45 | boss | authorize coder-agent:T3 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-17 00:15:22 | agent | coder-agent | ✅ scope-request · T3 · 36.7s · $0.3593 |
| 2026-10-17 00:15:22 | gate | gate-question | ❌ REJECTED: yes, grant internal/compile/refs.go and jsonschema.go for this run (the meta-schema check runs from the root package because eval imports compile). |
| 2026-10-17 00:15:37 | boss | authorize coder-agent:T3 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-17 00:18:54 | agent | coder-agent | ✅ complete · T3 · 196.9s · $0.8048 |
| 2026-10-17 00:19:01 | gate | gate-build | ✅ [T3] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 00:19:01 | boss | authorize coder-agent:T4 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-17 00:20:41 | agent | coder-agent | ✅ complete · T4 · 100.0s · $0.5764 |
| 2026-10-17 00:20:48 | gate | gate-build | ✅ [T4] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 00:20:48 | boss | authorize coder-agent:T5 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-17 00:21:44 | agent | coder-agent | ✅ complete · T5 · 55.0s · $0.4632 |
| 2026-10-17 00:21:52 | gate | gate-build | ✅ [T5] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 00:21:53 | boss | authorize tester-agent | ✅ gate-build passed with all 5 task(s) implemented |
| 2026-10-17 00:23:20 | agent | tester-agent | ✅ warn · 86.8s · $0.8150 |
| 2026-10-17 00:23:20 | gate | gate-test | ✅ QA/security/performance passed (with warnings). AC covered: 10; findings: 5 new, 0 persistent, 0 resolved |
| 2026-10-17 00:23:20 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-17 00:24:11 | agent | release-agent | ✅ warn · 51.3s · $0.0941 |
| 2026-10-17 00:24:12 | gate | gate-release | ❌ REJECTED: Fix these, nothing else: (F1) any failure of a reached registry schema or custom meta-schema in compileChecked is reported as InvalidSchemaError naming the file/URI and location, never quoting schema or inst… |
| 2026-10-17 00:25:24 | boss | authorize coder-agent:remediation | ✅ the operator's rejection at Checkpoint 3 (1 finding(s)) |
| 2026-10-17 00:30:33 | agent | coder-agent | ✅ scope-request · remediation · 307.8s · $2.3341 |
| 2026-10-17 00:30:33 | gate | gate-question | ❌ REJECTED: yes, grant registry_options.go, testdata/validation/maxItems.json and testdata/validation/minItems.json for this run (negative/fractional minItems/maxItems are invalid schemas in 2020-12 and belong in the ex… |
| 2026-10-17 00:30:54 | boss | authorize coder-agent:remediation | ✅ the operator's rejection at Checkpoint 3 (3 finding(s)) |
| 2026-10-17 00:35:39 | agent | coder-agent | ✅ complete · remediation · 283.6s · $2.3221 |
| 2026-10-17 00:35:54 | gate | gate-build | ✅ [remediation] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 00:35:55 | boss | authorize tester-agent | ✅ gate-build passed with all 5 task(s) implemented |
| 2026-10-17 00:36:44 | agent | tester-agent | ✅ warn · 49.4s · $0.7442 |
| 2026-10-17 00:36:44 | gate | gate-test | ✅ QA/security/performance passed (with warnings). AC covered: 10; findings: 3 new, 0 persistent, 0 resolved |
| 2026-10-17 00:36:44 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-17 00:37:31 | agent | release-agent | ✅ warn · 46.5s · $0.1008 |
| 2026-10-17 00:37:32 | gate | gate-release | ⏸️ awaiting human sign-off |
| 2026-10-17 00:37:32 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
