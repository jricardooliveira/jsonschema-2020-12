# Pipeline record — #14: Annotation collection: metadata, format, content and unknown keywords, evaluated properties and items

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #144 · **Status:** waiting_human · **Stage:** gate-2-human
- **Started:** 2026-10-16 21:28:36 UTC · **Stopped:** 2026-10-16 21:30:40 UTC
- **Story spend:** ~$0.98 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-2-architect: `factory approve 144` to accept, or `factory reject 144 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-2-architect**:
  > ⚠️ BREAKING CHANGES detected — these will affect existing API consumers:
  >   • A passing 'basic' report now lists annotation units, so it is longer than before. The operator approved this.
  >   • A passing 'basic' report now lists annotation units, so it is longer than before. The operator approved this. The NDJSON wrapper, the failing-report shape and the flag output do not change.
  >   → Approve these breaking changes?
  >
  > 🧱 BOUNDARY REVIEW WARNINGS — tenant / authorization / API contract / security concerns found before any code is written:
  >   • security: The design bounds annotations by count only (10,000,000). It does not bound the size of the report. An annotation value is a schema value that can be large (a big 'default', 'examples' or unknown-keyword value). A hostile schema, which the brief treats as possible, can attach one such value to a subschema applied to every item of a large document. The passing basic report then repeats the value once per instance location. In memory the values are shared by reference, so only the serialized report and CLI stdout grow, up to count times value size. The operator confirmed the count limit but never ruled on this output amplification. A human should confirm that the count limit alone is enough, or choose a bound on the total bytes of annotation values or of the rendered report, using the same LimitError style with exit 2.
  >   → Approve the design with these boundary risks, or reject with the fix?

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-16 21:28:36 | run | run started | • |
| 2026-10-16 21:28:37 | agent | spec-agent | ✅ pass |
| 2026-10-16 21:28:37 | gate | gate-1-spec | ✅ Story 'Annotation collection: metadata, format, content and unknown keywords, evaluated properties and items' has 11 AC and 6 tasks |
| 2026-10-16 21:28:37 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-16 21:28:37 | agent | architect-agent | ✅ pass |
| 2026-10-16 21:28:38 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security) |
| 2026-10-16 21:29:11 | agent | boundary-agent | ✅ warn · 33.6s · $0.2020 |
| 2026-10-16 21:29:11 | gate | gate-2-architect | ❌ REJECTED: Keep the design and settle the boundary points: (1) MaxAnnotations follows the existing limit style exactly like MaxOutputUnits: unset/0 in the library = DefaultMaxAnnotations, negative = compile error, expl… |
| 2026-10-16 21:29:38 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-16 21:30:21 | agent | architect-agent | ✅ pass · 42.5s · $0.5745 |
| 2026-10-16 21:30:21 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security) |
| 2026-10-16 21:30:40 | agent | boundary-agent | ✅ warn · 18.5s · $0.2005 |
| 2026-10-16 21:30:40 | gate | gate-2-architect | ⏸️ awaiting human sign-off |
| 2026-10-16 21:30:40 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
