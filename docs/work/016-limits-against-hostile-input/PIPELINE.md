# Pipeline record — #16: Limits against hostile input and proof of the load-once guarantee

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #152 · **Status:** waiting_human · **Stage:** gate-2-human
- **Started:** 2026-10-17 01:41:14 UTC · **Stopped:** 2026-10-17 01:51:10 UTC
- **Story spend:** ~$3.26 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-2-architect: `factory approve 152` to accept, or `factory reject 152 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-2-architect**:
  > ⚠️ BREAKING CHANGES detected — these will affect existing API consumers:
  >   • Public functions now return a top-level *jsonschema.LimitError for every exceeded limit, instead of a *jsonvalue.LimitError wrapped in 'schema:', 'document:' or 'registry:' text.
  >   • Internal limit names change: 'schema file bytes' becomes 'input size', 'total registry bytes' becomes 'registry size', and 'evaluation depth' becomes 'nesting depth', which carries the depth that was set.
  >   • An oversize pattern becomes a LimitError ('pattern size'), where it was a plain schema error with different wording.
  >   • CLI stderr for every limit becomes 'limit exceeded: <limit> <value> (raise with --<flag>)', replacing '<limit> exceeds limit <n> ...'. parse_errors_test.go is updated.
  >   • Registry.MaxFileBytes = 0 now follows Options.MaxInputBytes. The default value is unchanged (64 MiB).
  >   • Evaluation step counts rise, because steps are now charged per keyword and in proportion to the bytes compared or matched. A caller who set a low custom MaxSteps may hit it sooner.
  >   • Every exceeded limit now comes back from Compile, CompileWithOptions, CompileWithRegistry, Validate and Result.Output as a top-level *jsonschema.LimitError, not a wrapped internal error.
  >   • Limit names change: 'schema file bytes' becomes 'input size', 'total registry bytes' becomes 'registry size', and 'evaluation depth' becomes 'nesting depth' (carrying the configured depth).
  >   • An oversize pattern becomes a LimitError ('pattern size') instead of a plain schema error.
  >   • The CLI stderr line for every limit becomes 'limit exceeded: <limit> <value> (raise with --<flag>)', and number-cap errors lose their line and column.
  >   • Registry.MaxFileBytes = 0 now follows the resolved Options.MaxInputBytes (same 64 MiB default).
  >   • Step counts rise because charging is now per keyword and scaled by bytes, so a caller with a low custom MaxSteps may hit the cap sooner.
  >   • A pure $ref/$dynamicRef cycle now matches errors.As for LimitError through LoopError.Unwrap.
  >   → Approve these breaking changes?
  >
  > 🧱 BOUNDARY REVIEW WARNINGS — tenant / authorization / API contract / security concerns found before any code is written:
  >   • security: Work-step charging has gaps against operator decision 5 ('charged by all keywords'). The design charges one step per keyword in the dispatch, plus proportional charges only for uniqueItems, enum, const, contains, patternProperties, pattern, minLength and maxLength. Other keywords loop over schema-sized or instance-sized data. They are dependentRequired, dependentSchemas, the required name list, and every `format` checker in assertion mode, including `format: regex` on an instance-supplied string. A crafted 64 MiB schema can reference such a keyword many times through fan-out ($ref inside allOf or anyOf). Each hop costs one step but a long loop. A human should settle whether one dispatch step per keyword is enough, or whether each of these loops must charge in proportion. The rules below require proportional charging.
  >   • security: The pure-cycle LimitError produced through LoopError.Unwrap does not say which Value it carries. The binding rule is that it must be the configured MaxRefChain, and the design does not contradict it. The rule is kept.
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
| 2026-10-17 01:41:41 | gate | gate-2-architect | ❌ REJECTED: Keep the design; settle: (1) the boundary 'rules for the coder' are binding (bounded reads before decoding, depth checks, ref-chain per path, steps charged proportionally in super-linear keyword loops, fresh… |
| 2026-10-17 01:42:08 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-17 01:47:24 | agent | architect-agent | ✅ pass · 315.0s · $1.6910 |
| 2026-10-17 01:47:24 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security) |
| 2026-10-17 01:48:12 | agent | boundary-agent | ❌ fail · 48.0s · $0.2485 |
| 2026-10-17 01:48:12 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-17 01:50:06 | agent | architect-agent | ✅ pass · 113.0s · $0.8433 |
| 2026-10-17 01:50:07 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security) |
| 2026-10-17 01:51:10 | agent | boundary-agent | ✅ warn · 63.2s · $0.2823 |
| 2026-10-17 01:51:10 | gate | gate-2-architect | ⏸️ awaiting human sign-off |
| 2026-10-17 01:51:10 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
