# INTENT #11

- Author: Joao Oliveira
- Date: 2026-10-16
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Add 2020-12 '$dynamicRef' and '$dynamicAnchor' with dynamic scope, as the core specification describes. A $dynamicRef whose initial target holds a matching $dynamicAnchor resolves to the outermost matching $dynamicAnchor in the dynamic scope. Otherwise it behaves like $ref. This works across embedded resources and across schemas supplied through the registry. Write spec-derived positive and negative tests, including recursive extension of a schema.

## Operator clarifications
- Operator note on refining again — Settle the two open cases: per 2020-12 core 8.2.2, $dynamicAnchor also defines a plain-name fragment in the same namespace as $anchor, so a $ref to '#name' can target a $dynamicAnchor, and $dynamicRef to '#name' behaves like $ref when the initially resolved target has no matching $dynamicAnchor (bookending). A duplicate plain name within one schema resource (two $dynamicAnchor, or a $dynamicAnchor and an $anchor with the same name) is a broken schema, exit 2, consistent with story 186. Also make it an acceptance criterion that errors reached through $dynamicRef carry a keywordLocation containing '/$dynamicRef' and the absoluteKeywordLocation of the resolved target. Everything else is accepted as stated.
- Operator note on refining again — Do not leave $dynamicRef recursion unbounded until story 193. Story 186 adds evaluation-time loop detection (re-entering the same schema location at the same instance location on the current evaluation path is a loop error, exit 2); $dynamicRef must go through the same check, with a test of a non-terminating $dynamicRef cycle that ends with exit 2 and no hang. Story 193 later adds the configurable recursion limit on top. Everything else in this revision is accepted.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The repository inventory is empty, so the exact names of story 186's files are assumed. That covers the anchor index (assumed internal/registry/anchor.go), the $ref keyword (internal/keywords/core/ref.go), the loop check and the resource-entry point in internal/eval/eval.go. The coder may need to adjust these paths. If the duplicate-name check lives elsewhere, that file needs a write claim instead.
- The test runner's package and file naming for testdata/core/*.json are assumed from the conventions. The spec test file may belong next to the shared runner instead.
- Recursive-extension tests assume that type (179), properties/required (183) and items (184) exist to express the tree constraints. The error-location criterion assumes story 180's basic output. If tests are limited to boolean schemas these dependencies could drop, but the cases would be much weaker.
- Whether the core vocabulary package is named internal/keywords/core is assumed from the vocabulary-per-package convention.

## Proposed outcome

See `SPEC.md` in this directory — the spec-agent's acceptance criteria are
the machine-readable statement of the outcome, derived from the problem above.

## Affected users and systems

See `PLAN.md` → *Files that change*, and the project's `project-spec.json`
for the systems in scope.

## Constraints

The project's `PROJECT_RULES.md` and `project-spec.json` (stack, conventions,
NFRs and the `forbidden` list) are injected into every agent and bind this work.

## Open questions

Recorded in `SPEC.md` → *Open questions*. A non-empty list parks the run at
Checkpoint 1 for the operator rather than failing it.
