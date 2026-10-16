# Pipeline record — #14: Annotation collection: metadata, format, content and unknown keywords, evaluated properties and items

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #144 · **Status:** waiting_human · **Stage:** gate-2-human
- **Started:** 2026-10-16 21:28:36 UTC · **Stopped:** 2026-10-16 21:29:11 UTC
- **Story spend:** ~$0.20 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-2-architect: `factory approve 144` to accept, or `factory reject 144 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-2-architect**:
  > ⚠️ BREAKING CHANGES detected — these will affect existing API consumers:
  >   • A passing 'basic' report now lists annotation units, so it is longer than before (operator-approved).
  >   → Approve these breaking changes?
  >
  > 🧱 BOUNDARY REVIEW WARNINGS — tenant / authorization / API contract / security concerns found before any code is written:
  >   • api_contract: Options.MaxAnnotations 'must be positive', but a Go zero value is 0. The design does not say whether an unset field means 'use DefaultMaxAnnotations' or is a usage error. An explicit 0 or a negative value must be rejected, and an unset field must not silently turn collection off or reject every default Options. The existing limit style (stories 178/179, MaxOutputUnits) must be followed and the choice stated.
  >   • api_contract: The sort key (instance location, then keyword location) can tie, for example when the same subschema is reached through different dynamic paths. The design does not give a tie-break or say whether the comparison is by the pointer string or component by component, so 'exactly the same list' is not yet guaranteed.
  >   • api_contract: The passing 'basic' report grows annotation units. This is operator-approved. The multi-document NDJSON wrapper and the failing-report shape must not change.
  >   • security: The design says annotation values come only from the schema and no instance value appears in any report. But the properties, patternProperties and additionalProperties annotations hold the property NAMES that were evaluated, and additionalProperties lists exactly the names the schema did not mention. These come from the document, and keys can be sensitive (user ids, emails). If they are placed in Result.Annotations() and in the passing basic report, the 'schema values only' rule is broken. The operator should settle whether instance property names count as locations (already reported) or as instance data. Indices from prefixItems and contains are not a concern.
  >   • security: The 10,000,000 limit counts annotations, not bytes. An annotation value is a schema value that may be large (a big 'examples', 'default' or unknown-keyword value). With always-collect, a hostile schema could multiply that value across millions of instance locations, so memory and the size of the basic report are not bounded by the count. A human should confirm that the count limit alone is enough. Otherwise a size bound is needed.
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
| 2026-10-16 21:29:11 | gate | gate-2-architect | ⏸️ awaiting human sign-off |
| 2026-10-16 21:29:11 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
