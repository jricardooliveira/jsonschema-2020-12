# Pipeline record — #18: Format assertion: mode, dates, email, hostnames, IP and UUID

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #153 · **Status:** waiting_human · **Stage:** gate-release-human
- **Started:** 2026-10-17 02:32:34 UTC · **Stopped:** 2026-10-17 03:03:17 UTC
- **Story spend:** ~$8.51 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-release: `factory approve 153` to accept, or `factory reject 153 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-release**:
  > 🚀 RELEASE SIGN-OFF (Checkpoint 3) — NOT READY: 1 gap(s) in the evidence:
  >   • Security finding (medium) not resolved: None blocking. Format errors carry only the locations and a fixed message with the format name from the schema, never the instance value. Every check is hand-written, single-pass and linear, with no regexp and no net/netip/mail parsers. Each check is charged with ctx.ChargeBytes(len(s)) before it runs, and TestFormatAssertionChargesSteps shows this. The registry is written only from init(). The only fail-open paths are the specified ones: an unknown or deferred format, a non-string instance, or a non-string format value.
  >   ◦ Release-agent note: F1 (medium): The guard forcing meta-schema validation to annotation-only (noFormat=true in metacheck.go) is exercised by formatassertion_test.go::TestFormatAssertionMetaCheckAnnotationOnly, but the test only verifies compile succeeds. It does not fail without the guard; it would pass if the guard were removed and the meta-schema itself did not assert format-assertion on keywords like $id. An explicit test verifying the guard's effect (e.g., mocking or inspecting the compile path) would strengthen assurance that meta-checks stay annotation-only in all modes.
  >   ◦ Release-agent note: F3 (low): No test covers an embedded resource with a different $schema than the root, which could have a different format-assertion vocabulary flag recorded by story 190. This edge case is not in the acceptance criteria but is a potential interaction point.
  >   ◦ Release-agent note: F2 (low): IPv6 address literals in email must use the 'IPv6:' prefix per RFC 5321 section 4.1.3; user@[::1] is invalid, user@[IPv6:::1] is valid. This is standard-conformant but differs from some real-world email parsers that tolerate the bare form.
  >   → Approve to release accepting the gaps above as a known risk, or reject with what must change — it goes back to the coder, then the tester.

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-17 02:32:34 | run | run started | • |
| 2026-10-17 02:32:34 | agent | spec-agent | ✅ pass |
| 2026-10-17 02:32:34 | gate | gate-1-spec | ✅ Story 'Format assertion: mode, dates, email, hostnames, IP and UUID' has 17 AC and 6 tasks |
| 2026-10-17 02:32:35 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-17 02:32:35 | agent | architect-agent | ✅ pass |
| 2026-10-17 02:32:35 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it touches sensitive areas (security) |
| 2026-10-17 02:32:53 | agent | boundary-agent | ✅ pass · 18.2s · $0.1732 |
| 2026-10-17 02:32:53 | gate | gate-2-architect | ✅ Architecture approved. 32 modules, 3 risks flagged |
| 2026-10-17 02:32:53 | boss | authorize coder-agent:T1 | ✅ gate-2-architect passed |
| 2026-10-17 02:33:26 | agent | coder-agent | ✅ scope-request · T1 · 32.1s · $0.3477 |
| 2026-10-17 02:33:26 | gate | gate-question | ❌ REJECTED: yes, grant internal/keywords/registry.go, internal/compile/refs.go, jsonschema.go and numberlimits.go for this run (wiring the format-assertion option; no other story is running). The flag fixed at compile t… |
| 2026-10-17 02:33:40 | boss | authorize coder-agent:T1 | ✅ gate-2-architect passed |
| 2026-10-17 02:35:03 | agent | coder-agent | ✅ complete · T1 · 82.6s · $0.4816 |
| 2026-10-17 02:36:42 | gate | gate-build | ✅ [T1] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 02:36:43 | boss | authorize coder-agent:T2 | ✅ gate-2-architect passed |
| 2026-10-17 02:37:34 | agent | coder-agent | ✅ complete · T2 · 51.2s · $0.4125 |
| 2026-10-17 02:39:09 | gate | gate-build | ✅ [T2] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 02:39:09 | boss | authorize coder-agent:T4 | ✅ gate-2-architect passed |
| 2026-10-17 02:39:38 | agent | coder-agent | ✅ complete · T4 · 28.6s · $0.3446 |
| 2026-10-17 02:41:14 | gate | gate-build | ✅ [T4] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 02:41:14 | boss | authorize coder-agent:T3 | ✅ gate-2-architect passed |
| 2026-10-17 02:42:20 | agent | coder-agent | ✅ complete · T3 · 65.3s · $0.4621 |
| 2026-10-17 02:43:56 | gate | gate-build | ✅ [T3] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 02:43:57 | boss | authorize coder-agent:T5 | ✅ gate-2-architect passed |
| 2026-10-17 02:46:35 | agent | coder-agent | ✅ scope-request · T5 · 158.4s · $0.6010 |
| 2026-10-17 02:46:35 | gate | gate-question | ❌ REJECTED: yes, grant testdata/vocabulary/required_optional_unknown.json for this run: story 190's interim behaviour (format-assertion vocabulary required but not asserted) is replaced by this story's required-vocabula… |
| 2026-10-17 02:46:51 | boss | authorize coder-agent:T5 | ✅ gate-2-architect passed |
| 2026-10-17 02:48:31 | agent | coder-agent | ✅ complete · T5 · 100.2s · $0.5076 |
| 2026-10-17 02:48:55 | gate | gate-build | ✅ [T5] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 02:48:55 | boss | authorize coder-agent:T6 | ✅ gate-2-architect passed |
| 2026-10-17 02:49:10 | agent | coder-agent | ✅ complete · T6 · 14.2s · $0.3477 |
| 2026-10-17 02:49:10 | gate | gate-build | ✅ [T6] go_test:pass, go_race:pass |
| 2026-10-17 02:49:11 | boss | authorize tester-agent | ✅ gate-build passed with all 6 task(s) implemented |
| 2026-10-17 02:50:39 | agent | tester-agent | ✅ warn · 88.4s · $0.7989 |
| 2026-10-17 02:50:39 | gate | gate-test | ✅ QA/security/performance passed (with warnings). AC covered: 17; findings: 3 new, 0 persistent, 0 resolved |
| 2026-10-17 02:50:39 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-17 02:51:37 | agent | release-agent | ✅ warn · 57.6s · $0.1029 |
| 2026-10-17 02:51:38 | gate | gate-release | ❌ REJECTED: Fix only: (F1) internal/formats/duration.go must follow the RFC 3339 Appendix A ABNF exactly: dur-year = 1*DIGIT "Y" [dur-month]; dur-month = 1*DIGIT "M" [dur-day]; dur-day = 1*DIGIT "D"; dur-hour = 1*DIGIT … |
| 2026-10-17 02:51:59 | boss | authorize coder-agent:remediation | ✅ the operator's rejection at Checkpoint 3 (1 finding(s)) |
| 2026-10-17 02:54:50 | agent | coder-agent | ✅ complete · remediation · 170.3s · $1.1846 |
| 2026-10-17 02:55:13 | gate | gate-build | ✅ [remediation] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 02:55:14 | boss | authorize tester-agent | ✅ gate-build passed with all 6 task(s) implemented |
| 2026-10-17 02:56:19 | agent | tester-agent | ✅ warn · 65.5s · $0.7173 |
| 2026-10-17 02:56:19 | gate | gate-test | ✅ QA/security/performance passed (with warnings). AC covered: 17; findings: 3 new, 0 persistent, 0 resolved |
| 2026-10-17 02:56:19 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-17 02:57:33 | agent | release-agent | ✅ warn · 73.4s · $0.1112 |
| 2026-10-17 02:57:33 | gate | gate-release | ❌ REJECTED: Fix only: (F1) in internal/formats/email.go match the RFC 5321 address-literal tag 'IPv6:' case-insensitively (ABNF literals are case-insensitive), test user@[ipv6:::1] and user@[IPV6:::1] valid; (F3) add a … |
| 2026-10-17 02:58:15 | boss | authorize coder-agent:remediation | ✅ the operator's rejection at Checkpoint 3 (1 finding(s)) |
| 2026-10-17 02:59:08 | agent | coder-agent | ✅ complete · remediation · 52.6s · $0.9692 |
| 2026-10-17 03:00:46 | gate | gate-build | ✅ [remediation] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-17 03:00:47 | boss | authorize tester-agent | ✅ gate-build passed with all 6 task(s) implemented |
| 2026-10-17 03:02:15 | agent | tester-agent | ✅ warn · 88.6s · $0.8434 |
| 2026-10-17 03:02:15 | gate | gate-test | ✅ QA/security/performance passed (with warnings). AC covered: 17; findings: 3 new, 0 persistent, 0 resolved |
| 2026-10-17 03:02:16 | boss | authorize release-agent | ✅ gate-test passed |
| 2026-10-17 03:03:16 | agent | release-agent | ✅ warn · 60.5s · $0.1021 |
| 2026-10-17 03:03:17 | gate | gate-release | ⏸️ awaiting human sign-off |
| 2026-10-17 03:03:17 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
