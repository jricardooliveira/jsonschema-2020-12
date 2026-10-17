# Pipeline record — #16: Limits against hostile input and proof of the load-once guarantee

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #152 · **Status:** waiting_human · **Stage:** gate-2-human
- **Started:** 2026-10-17 01:41:14 UTC · **Stopped:** 2026-10-17 01:41:41 UTC
- **Story spend:** ~$0.19 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-2-architect: `factory approve 152` to accept, or `factory reject 152 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-2-architect**:
  > ⚠️ BREAKING CHANGES detected — these will affect existing API consumers:
  >   • Errors from existing number/depth caps become *jsonschema.LimitError; stderr wording for oversize inputs changes to the unified format.
  >   • Errors from the existing number-digit, number-exponent, depth and input-size caps become *jsonschema.LimitError.
  >   • The stderr wording for oversize inputs changes to 'limit exceeded: <limit> <value> (raise with --<flag>)'.
  >   • If any existing limit flag is renamed, that is a CLI-breaking change. Keep the old names working.
  >   → Approve these breaking changes?
  >
  > 🧱 BOUNDARY REVIEW WARNINGS — tenant / authorization / API contract / security concerns found before any code is written:
  >   • api_contract: api_design says the existing number, pattern and annotation CLI flags are 'renamed only if they do not already match'. A rename would break the published CLI contract and TestNumberCapFlags. The design does not say whether the old names stay as aliases. A human should settle this.
  >   • api_contract: ADR-014 left the report-byte cap (Options.MaxReportBytes, DefaultMaxReportBytes) without a CLI flag and assigned that flag to story 193. The design's Limits list, Default* constants, flag list and LimitError names do not include it. The cap would then stay library-only and outside the unified LimitError style. Add it, or have the operator defer it explicitly.
  >   • api_contract: LimitError{Limit string; Value int64} has no fixed, documented set of Limit names. Callers branch on these names, and the CLI maps them to flag names. The names must be a closed, exported set, since they become public API.
  >   • security: The work-step counter is described as a single increment per checking step. The design does not say it charges keywords whose cost is super-linear in one instance, such as uniqueItems (pairwise deep equality), contains, patternProperties × properties, enum/const deep comparison, and regex matching over large strings. A crafted 64 MiB document could then hang the tool inside one keyword without reaching 100M steps. The step charge must scale with that internal work.
  >   • security: The design does not state the order of the size check and the read. If a file or stdin is read fully before the size check, a hostile input exhausts memory first, and the registry total check has the same problem.
  >   • security: Meta-schema validation at compile time runs the evaluator. The design does not say whether it uses the step and ref-chain budget, or whether a LimitError from it can be wrapped as InvalidSchemaError. That would turn a limit error into a schema error and break the 'LimitError is never a schema error' rule.
  >   • security: The ADR-017 detailed/verbose render re-evaluates the instance inside Result.Output. The design does not say that this second evaluation gets its own fresh budget with the same limits.
  >   → Approve the design with these boundary risks, or reject with the fix?

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-17 01:41:14 | run | run started | • |
| 2026-10-17 01:41:14 | agent | spec-agent | ✅ pass |
| 2026-10-17 01:41:14 | gate | gate-1-spec | ✅ Story 'Limits against hostile input and proof of the load-once guarantee' has 11 AC and 6 tasks |
| 2026-10-17 01:41:15 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-17 01:41:15 | agent | architect-agent | ✅ pass |
| 2026-10-17 01:41:15 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security) |
| 2026-10-17 01:41:41 | agent | boundary-agent | ✅ warn · 26.3s · $0.1910 |
| 2026-10-17 01:41:41 | gate | gate-2-architect | ⏸️ awaiting human sign-off |
| 2026-10-17 01:41:41 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
