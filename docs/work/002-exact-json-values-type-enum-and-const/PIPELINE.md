# Pipeline record — #2: Exact JSON value model, deep equality, and the type, enum and const keywords

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #135 · **Status:** waiting_human · **Stage:** gate-2-human
- **Started:** 2026-10-16 16:50:29 UTC · **Stopped:** 2026-10-16 16:56:37 UTC
- **Story spend:** ~$1.25 of $26.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-2-architect: `factory approve 135` to accept, or `factory reject 135 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-2-architect**:
  > ⚠️ BREAKING CHANGES detected — these will affect existing API consumers:
  >   • The `type` keyword is now enforced. Schemas that used it were previously accepted as unknown-keyword no-ops, which changes 178's TestValidateSchemas expectation for {"type":"string"}.
  >   • Input that 178 accepted is now rejected as broken JSON (exit 2): invalid UTF-8, unpaired \u surrogates, and numbers over the digit or exponent cap. A leading BOM is now accepted (178 rejected it).
  >   • `type` is now enforced. Schemas that used it were previously ignored as an unknown keyword.
  >   • Input that story 178 accepted is now rejected with exit 2: invalid UTF-8, unpaired \u surrogates, and numbers over the digit or exponent cap.
  >   • A leading UTF-8 BOM, which 178 rejected, is now accepted.
  >   → Approve these breaking changes?
  >
  > 🧱 BOUNDARY REVIEW WARNINGS — tenant / authorization / API contract / security concerns found before any code is written:
  >   • api_contract: Checkpoint 2 decision 5 says every story 178 test keeps passing unchanged. The design edits one of them: the {"type":"string"} entry in jsonschema_test.go TestValidateSchemas, which expects an unknown-keyword no-op. Implementing `type` makes that expectation wrong, so the conflict is unavoidable. The code must not override an operator decision, so the operator should confirm the one-entry change (replace it with an unknown keyword such as x-unknown) or choose another route.
  >   • api_contract: The operator decision names new Reason values for the number-size case. The design uses the existing *LimitError with new Limit names ("number digits", "number exponent"). This keeps the typed errors.As mapping and adds no ParseError family, so it honours the intent. The operator should know the wording differs.
  >   • api_contract: In the library a zero Options field means the default and a negative value is an error. The operator wording is 'zero or negative = usage error'. Only the CLI can tell an explicit 0 from an absent flag, and there an explicit 0 is a usage error. A zero library field is never 'no limit', so this is safe, but it is an interpretation.
  >   • security: CPU amplification on hostile input. Numbers are re-parsed into big.Rat on every Equal call, and each number may carry up to the default cap of digits and exponent. A large document of cap-sized numbers compared against many enum or const values (or nested arrays) multiplies the number of expensive parses. Input size is capped, but evaluation steps are not capped until story 193. The design defers caching to 'if profiling shows a need', which leaves the hostile-input requirement open. A human should confirm that cheap fast paths are required in this story and that story 193 counts Equal work against the evaluation-step limit.
  >   • security: Parser differential risk. The hand-written pre-scan (UTF-8, escapes, surrogates) and encoding/json's tokenizer must agree on where strings begin and end. Any disagreement could hide a lone surrogate or invalid byte from the check. The design describes backslash handling but names no test that cross-checks the two.
  >   → Approve the design with these boundary risks, or reject with the fix?

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-16 16:50:29 | run | run started | • |
| 2026-10-16 16:50:29 | agent | spec-agent | ✅ pass |
| 2026-10-16 16:50:29 | gate | gate-1-spec | ✅ Story 'Exact JSON value model, deep equality, and the type, enum and const keywords' has 11 AC and 6 tasks |
| 2026-10-16 16:50:30 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-16 16:50:30 | agent | architect-agent | ✅ pass |
| 2026-10-16 16:50:30 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security) |
| 2026-10-16 16:51:03 | agent | boundary-agent | ✅ warn · 33.1s · $0.1646 |
| 2026-10-16 16:51:03 | gate | gate-2-architect | ❌ REJECTED: Keep the design (the internal evaluate-signature change to jsonvalue.Value is accepted: internal only) and settle every boundary point: (1) Errors: reuse ADR-001's typed *SyntaxError and *LimitError in inter… |
| 2026-10-16 16:53:14 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-16 16:55:55 | agent | architect-agent | ✅ pass · 160.3s · $0.8831 |
| 2026-10-16 16:55:55 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security) |
| 2026-10-16 16:56:37 | agent | boundary-agent | ✅ warn · 41.5s · $0.2033 |
| 2026-10-16 16:56:37 | gate | gate-2-architect | ⏸️ awaiting human sign-off |
| 2026-10-16 16:56:37 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
