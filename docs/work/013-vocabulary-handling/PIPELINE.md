# Pipeline record — #13: Honour $vocabulary from embedded and registry meta-schemas

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #151 · **Status:** waiting_human · **Stage:** gate-release-human
- **Started:** 2026-10-17 01:06:09 UTC · **Stopped:** 2026-10-17 01:23:37 UTC
- **Story spend:** ~$3.86 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-release: `factory approve 151` to accept, or `factory reject 151 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-release**:
  > 🚀 RELEASE SIGN-OFF (Checkpoint 3) — NOT READY: 1 gap(s) in the evidence:
  >   • Security finding (medium) not resolved: F1 (medium): errors from an externally referenced registry resource that declares its own '$schema' lose their type. They pass through the existing refs.go line 212 wrap, fmt.Errorf("...: %v", err). Examples are UnknownVocabularyError and CoreVocabularyError. The CLI still exits 2, so the check still fails closed, but errors.As cannot see the type, which falls short of AC2's library contract for that path. No test covers a $ref-reached resource with its own vocabulary set.
  >   ◦ Release-agent note: Breaking change: basic and detailed reports no longer contain '$schema', '$vocabulary' or '$comment' annotations. Tests and evaluators that expected these annotations will need updating.
  >   ◦ Release-agent note: The change routes inactive-vocabulary keywords through the existing unknown-keyword path in compile.go, a crowded file. The edit is surgical (four lines in the keyword loop), but future changes to that loop must be aware of the new condition.
  >   ◦ Release-agent note: Error message for malformed '$vocabulary' entries (the bound entry key) is asserted only at unit test level, not in CLI integration tests (F3). The message is generated and used correctly, but large entry keys longer than 80 characters are silently truncated in the error message.
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
| 2026-10-17 01:23:37 | gate | gate-release | ⏸️ awaiting human sign-off |
| 2026-10-17 01:23:37 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
