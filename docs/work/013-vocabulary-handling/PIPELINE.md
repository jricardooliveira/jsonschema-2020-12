# Pipeline record — #13: Honour $vocabulary from embedded and registry meta-schemas

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #151 · **Status:** waiting_human · **Stage:** gate-release-human
- **Started:** 2026-10-17 01:06:09 UTC · **Stopped:** 2026-10-17 01:31:49 UTC
- **Story spend:** ~$8.06 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-release: `factory approve 151` to accept, or `factory reject 151 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-release**:
  > 🚀 RELEASE SIGN-OFF (Checkpoint 3) — NOT READY: 1 gap(s) in the evidence:
  >   • Security finding (medium) not resolved: No insecure default found. ParseVocabulary fails closed: every malformed, core or unknown-required path returns an error, and only an absent '$vocabulary' means all vocabularies are active. The recognised set is the fixed list of eight standard URIs. The $schema chain is bounded by maxMetaChain and seen-set cycle detection. Meta-schemas are resolved only from the ext index (embedded plus caller registry), so nothing is fetched or read from disk. Error text carries URIs, a regex-derived draft name and an entry key bounded to 80 bytes, never instance values.
  >   ◦ Release-agent note: Test coverage gap (medium, AC8): the CLI test for a malformed '$vocabulary' entry checks only that the filename appears in the error, not that the entry details are included. The library test in vocabulary_test.go does verify the message format. Recommend: add an explicit CLI case that asserts the full message includes the bad entry (e.g. 'invalid '$vocabulary' entry "value-not-boolean": value is not a boolean')
  >   ◦ Release-agent note: Two test files outside the originally planned scope were updated: testdata/metaschema/schema_keyword.json and cmd/jsonschema/metaschema_cli_test.go. Both changes are sound (the draft-07 error message must match the new format), but a reviewer should confirm the expected message change does not conflict with story 189's earlier handling.
  >   ◦ Release-agent note: The malformed-vocabulary code correctly bounds long URIs to 80 bytes via bound(); callers detecting UnknownVocabularyError or MalformedVocabularyError will receive the bounded URI, not the original. This is intentional (keep error messages short and safe), but if an operator later needs the full URI in logs, the implementation would have to be revised.
  >   → Approve to release accepting the gaps above as a known risk, or reject with what must change — it goes back to the coder, then the tester.

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-17 01:06:09 | run | run started | • |
| 2026-10-17 01:06:10 | agent | spec-agent | ✅ pass |
| 2026-10-17 01:06:10 | gate | gate-1-spec | ✅ Story 'Honour $vocabulary from embedded and registry meta-schemas' has 12 AC and 5 tasks |
| 2026-10-17 01:06:10 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-17 01:06:10 | agent | architect-agent | ✅ pass |
| 2026-10-17 01:06:11 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security) |
| 2026-10-17 01:06:36 | agent | boundary-agent | ✅ pass · 25.8s · $0.1882 |
| 2026-10-17 01:06:36 | gate | gate-2-architect | ✅ APPROVED: Approved by human reviewer |
| 2026-10-17 01:06:50 | boss | authorize coder-agent:T1 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-17 01:07:36 | agent | coder-agent | ✅ complete · T1 · 45.6s · $0.3208 |
| 2026-10-17 01:09:12 | gate | gate-build | ✅ [T1] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 01:09:12 | boss | authorize coder-agent:T4 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-17 01:09:59 | agent | coder-agent | ✅ complete · T4 · 46.0s · $0.3437 |
| 2026-10-17 01:11:34 | gate | gate-build | ✅ [T4] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 01:11:34 | boss | authorize coder-agent:T2 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-17 01:12:35 | agent | coder-agent | ✅ scope-request · T2 · 60.7s · $0.3975 |
| 2026-10-17 01:12:35 | gate | gate-question | ❌ REJECTED: yes, grant testdata/metaschema/schema_keyword.json for this run (the message may name the schema's own $schema value; it is schema content, never instance data). |
| 2026-10-17 01:12:51 | boss | authorize coder-agent:T2 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-17 01:13:49 | agent | coder-agent | ✅ complete · T2 · 56.9s · $0.3707 |
| 2026-10-17 01:15:23 | gate | gate-build | ✅ [T2] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 01:15:23 | boss | authorize coder-agent:T3 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-17 01:15:53 | agent | coder-agent | ✅ scope-request · T3 · 29.7s · $0.3407 |
| 2026-10-17 01:15:53 | gate | gate-question | ❌ REJECTED: yes, grant internal/compile/refs.go and jsonschema.go for this run (the production keyword loop is in refs.go; the format-assertion record accessor lives on Schema). No other story is running. |
| 2026-10-17 01:16:14 | boss | authorize coder-agent:T3 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-17 01:17:54 | agent | coder-agent | ✅ complete · T3 · 99.8s · $0.4751 |
| 2026-10-17 01:19:26 | gate | gate-build | ✅ [T3] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 01:19:26 | boss | authorize coder-agent:T5 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-17 01:21:30 | agent | coder-agent | ✅ complete · T5 · 123.5s · $0.5775 |
| 2026-10-17 01:21:52 | gate | gate-build | ✅ [T5] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 01:21:52 | boss | authorize tester-agent | ✅ gate-build passed with all 5 task(s) implemented |
| 2026-10-17 01:23:04 | agent | tester-agent | ✅ warn · 71.8s · $0.7703 |
| 2026-10-17 01:23:04 | gate | gate-test | ✅ QA/security/performance passed (with warnings). AC covered: 12; findings: 3 new, 0 persistent, 0 resolved |
| 2026-10-17 01:23:04 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-17 01:23:36 | agent | release-agent | ✅ warn · 31.8s · $0.0778 |
| 2026-10-17 01:23:37 | gate | gate-release | ❌ REJECTED: Fix these two, nothing else: (F1) in internal/compile/refs.go compileDocument wrap target-compile errors with %w (not %v) so UnknownVocabularyError, unsupported-$schema and InvalidSchemaError from a $ref-rea… |
| 2026-10-17 01:24:10 | boss | authorize coder-agent:remediation | ✅ the operator's rejection at Checkpoint 3 (1 finding(s)) |
| 2026-10-17 01:25:16 | agent | coder-agent | ✅ complete · remediation · 66.4s · $1.1573 |
| 2026-10-17 01:25:39 | gate | gate-build | ✅ [remediation] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 01:25:39 | boss | authorize tester-agent | ✅ gate-build passed with all 5 task(s) implemented |
| 2026-10-17 01:26:56 | agent | tester-agent | ✅ warn · 76.7s · $0.7924 |
| 2026-10-17 01:26:56 | gate | gate-test | ✅ QA/security/performance passed (with warnings). AC covered: 12; findings: 5 new, 0 persistent, 0 resolved |
| 2026-10-17 01:26:56 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-17 01:27:28 | agent | release-agent | ✅ warn · 31.5s · $0.0794 |
| 2026-10-17 01:27:29 | gate | gate-release | ❌ REJECTED: Last narrow round, nothing else: (F3) add a test proving AC4: under a custom meta-schema listing only core+applicator, {"minimum":5} does not reject 3 AND '/minimum' appears as an annotation (value 5) in the… |
| 2026-10-17 01:27:56 | boss | authorize coder-agent:remediation | ✅ the operator's rejection at Checkpoint 3 (1 finding(s)) |
| 2026-10-17 01:29:16 | agent | coder-agent | ✅ complete · remediation · 78.6s · $1.2012 |
| 2026-10-17 01:29:46 | gate | gate-build | ✅ [remediation] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 01:29:46 | boss | authorize tester-agent | ✅ gate-build passed with all 5 task(s) implemented |
| 2026-10-17 01:31:01 | agent | tester-agent | ✅ warn · 74.1s · $0.8815 |
| 2026-10-17 01:31:01 | gate | gate-test | ✅ QA/security/performance passed (with warnings). AC covered: 12; findings: 3 new, 0 persistent, 0 resolved |
| 2026-10-17 01:31:01 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-17 01:31:48 | agent | release-agent | ✅ warn · 47.2s · $0.0908 |
| 2026-10-17 01:31:49 | gate | gate-release | ⏸️ awaiting human sign-off |
| 2026-10-17 01:31:49 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
