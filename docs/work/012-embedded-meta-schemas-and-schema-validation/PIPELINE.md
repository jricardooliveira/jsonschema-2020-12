# Pipeline record — #12: Embedded 2020-12 meta-schemas and meta-schema validation of every loaded schema

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #148 · **Status:** waiting_human · **Stage:** gate-release-human
- **Started:** 2026-10-17 00:11:06 UTC · **Stopped:** 2026-10-17 00:24:12 UTC
- **Story spend:** ~$4.10 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-release: `factory approve 148` to accept, or `factory reject 148 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-release**:
  > 🚀 RELEASE SIGN-OFF (Checkpoint 3) — NOT READY: 2 gap(s) in the evidence:
  >   • 1 acceptance criteria are not proven by any test the tester could see (§5.1) — verify them yourself or send the story back: ["The meta-schema check never asserts 'format', whether or not format-assertion mode is on, so enabling format assertion never makes a previously accepted schema fail."]
  >   • Security finding (medium) not resolved: F1 (medium) A failure in a supplied schema is not always reported as InvalidSchemaError, and that error path can quote schema values. In jsonschema.go compileChecked, when compile.CompileReached fails, only the main schema is re-run through metaCheck. A reached registry schema, or a custom meta-schema, that a keyword rejects at compile time (for example {"type":"nope"} or {"minLength":-1}) is returned as the plain error 'schema: ...'. That message carries the keyword compile text, such as validation/type.go 'type: unknown type name %q', which echoes the schema value. This breaks the boundary rule that meta-schema failures never quote schema values. It also breaks the rule that a reached registry schema failing its meta-schema is an InvalidSchemaError naming the file. The test 'custom meta-schema itself invalid' in custom_metaschema.json only asserts want:"error", so it passes in this state. No test covers a reached registry schema with a compile-level meta-schema violation.; F3 (low) The meta-check uses eval.DefaultMaxAnnotations instead of the caller's MaxAnnotations (jsonschema.go metaCheck). Step limits do apply, and limit errors stay plain errors with exit 2 (fail-closed). But the boundary rule asks for the same limits as normal evaluation, annotation count included.
  >   ◦ Release-agent note: AC7 (format assertion never asserts in meta-schema check) is not proven because format-assertion mode does not yet exist. This is a prerequisite gap: the feature must ship in a later story, and then this AC must be tested by verifying a schema with format errors still compiles.
  >   ◦ Release-agent note: When a reached registry schema's compilation fails (not meta-schema validation, but parsing or reference errors), the error is a plain error that may echo schema content. The error safety boundary for schema values applies to InvalidSchemaError only. Verify that ResolveMetaSchema and MetaChecks never echo schema values in their error messages (they don't: both say 'not supported' and 'does not end in 2020-12', never quoting URIs). Plain compilation errors from other parts of the pipeline are outside this story's value-free contract.
  >   ◦ Release-agent note: spec_test.go now silently skips test groups whose schema fails meta-validation. This weakens coverage of those spec sections (which can no longer be tested, since the schema does not compile). This is correct and expected: a schema that fails meta-validation is a load error, not a validation scenario. However, review existing test results to confirm no previously-passing spec group coverage is lost to an unexpected error.
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
| 2026-10-17 00:24:12 | gate | gate-release | ⏸️ awaiting human sign-off |
| 2026-10-17 00:24:12 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
