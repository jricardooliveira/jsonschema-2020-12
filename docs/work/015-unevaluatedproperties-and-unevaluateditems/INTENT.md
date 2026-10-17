# INTENT #15

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Add the 2020-12 'unevaluatedProperties' and 'unevaluatedItems' keywords, each in its own file and registered through the keyword registry. They use the collected annotations from every adjacent applicator and from nested ones: properties, patternProperties, additionalProperties, prefixItems, items, contains, allOf, anyOf, oneOf, not, if/then/else, dependentSchemas, $ref, $dynamicRef, and nested unevaluated keywords. Only successful subschemas count. Write spec-derived positive and negative tests, including cases that go through $ref and $dynamicRef.

## Operator clarifications
- When 'unevaluatedProperties' or 'unevaluatedItems' rejects a document because several properties or items were not covered by any other rule, how should the report list them? For example, if a document has two extra fields, 'nickname' and 'age', that no rule covers. — One error per extra field or item — Recommended. Each error points to that field or item's own location in the document, such as /nickname and /age. Scripts can see every offender directly. Reports still name only the location and the rule, never the value.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- Story 191 may already provide sibling-ordering or a phase mechanism. If so, the writes to internal/keywords/registry.go and internal/eval/eval.go (T1) are unnecessary and should be dropped. If it does not, T1 is the minimal wiring needed.
- The exact shape of the evaluated-properties/items annotations in EvalContext is defined by story 191 and not yet in the inventory; T2/T3 assume they are keyed by instance location and dropped on subschema failure.
- Story 185 must evaluate all anyOf/oneOf branches, without short-circuiting, for annotation completeness. If it short-circuits, that file needs a change, owned by 185 or by a follow-up story.
- Per-property/per-item error locations rely on the error/output-unit model from story 180; the current inventory's eval.Evaluate returns only (bool, error).
- The keyword file name convention '<keyword>.go' is applied in lower case (unevaluatedproperties.go); the testdata files keep the keyword's spelling.

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
