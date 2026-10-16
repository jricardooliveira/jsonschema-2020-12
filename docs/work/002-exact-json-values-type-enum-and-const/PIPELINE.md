# Pipeline record — #2: Exact JSON value model, deep equality, and the type, enum and const keywords

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #135 · **Status:** waiting_human · **Stage:** gate-2-human
- **Started:** 2026-10-16 16:50:29 UTC · **Stopped:** 2026-10-16 16:59:45 UTC
- **Story spend:** ~$2.10 of $26.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-2-architect: `factory approve 135` to accept, or `factory reject 135 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-2-architect**:
  > ⚠️ BREAKING CHANGES detected — these will affect existing API consumers:
  >   • The `type` keyword is now enforced, which changes 178's TestValidateSchemas expectation for {"type":"string"}; that entry becomes an unknown keyword.
  >   • Input that 178 accepted is now rejected as broken JSON (exit 2): invalid UTF-8, unpaired \u surrogates, and numbers over the digit or exponent cap. A leading BOM is now accepted (178 rejected it).
  >   • The `type` keyword is now enforced. The {"type":"string"} entry in 178's TestValidateSchemas becomes {"x-unknown":"string"}. The operator accepted this at Checkpoint 2.
  >   • Input that 178 accepted is now rejected with exit code 2: invalid UTF-8, unpaired \u surrogates, and numbers over the digit or exponent cap. The operator accepted this.
  >   • A leading BOM is now accepted. 178 rejected it.
  >   → Approve these breaking changes?
  >
  > 🧱 BOUNDARY REVIEW WARNINGS — tenant / authorization / API contract / security concerns found before any code is written:
  >   • security: Memory and CPU amplification from eagerly building a big.Rat for every number at decode time. This follows the operator's Fix 1, but it opens a new hostile-input path. With the default cap, a 7-byte token such as `1e10000` becomes a big.Rat holding 10^10000 (about 4 KB), and SetString has to compute that value. A 64 MiB document of such tokens (roughly 9 million numbers) would need tens of GB of memory and a lot of CPU. It would do so before any keyword runs and even when the schema has no numeric keyword. The risk entry calls this 'linear in input size and bounded', but the amplification factor is about 600x. It is not bounded by the input size cap. Equal, Cmp and IsInteger are cheap once the rats exist, so the decode step is now the cost. A human should confirm one of these: (a) a per-decode budget on the total expanded number size, with a LimitError when it is exceeded; (b) a compact exact form (big.Int mantissa plus an int exponent) that is expanded only when two numbers of very different scale must be compared; or (c) accepting the risk and recording it for story 193. Do not defer it with an unspecified 'later'.
  >   • security: Not a defect: the Fix 2 differential test (table plus fuzz target with f.Add seeds) now cross-checks the string boundaries between scanText and encoding/json, and checks that no U+FFFD can appear silently. The earlier parser-differential concern is closed. The earlier concern about re-parsing numbers in Equal is also closed, because Equal no longer parses.
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
| 2026-10-16 16:56:37 | gate | gate-2-architect | ❌ REJECTED: Approve everything in this revision except two security points, which must be fixed in this story. Explicitly accepted: replacing the {"type":"string"} entry of 178's TestValidateSchemas with an unknown keyw… |
| 2026-10-16 16:57:24 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-16 16:59:06 | agent | architect-agent | ✅ pass · 102.0s · $0.6455 |
| 2026-10-16 16:59:06 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security) |
| 2026-10-16 16:59:45 | agent | boundary-agent | ✅ warn · 38.5s · $0.2054 |
| 2026-10-16 16:59:45 | gate | gate-2-architect | ⏸️ awaiting human sign-off |
| 2026-10-16 16:59:45 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
