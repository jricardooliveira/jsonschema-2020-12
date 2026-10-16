# INTENT #8

- Author: Joao Oliveira
- Date: 2026-10-16
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Add the 2020-12 applicator keywords allOf, anyOf, oneOf, not, and if/then/else. Each keyword lives in its own file and registers through the keyword registry. 'then' and 'else' have no effect without 'if', and a failure inside 'if' alone never makes the document invalid. Boolean schemas true and false work as subschemas everywhere. Write spec-derived positive and negative tests, including nested combinations and oneOf with more than one match.

## Operator clarifications
- When a document fails 'anyOf' (none of the choices matched) or 'oneOf' (zero choices matched), what should the 'basic' report list? — The rule that failed plus each choice's own failures — Recommended. One entry saying anyOf/oneOf failed, followed by the reasons each choice did not match, each with its location and rule (never the data). Longer, but shows the user what to fix.
- When 'oneOf' fails because more than one choice matched, what should the report say? — Name which choices matched — Recommended. For example 'oneOf: choices 0 and 2 both matched', with the location and rule but no data. Shows the schema author exactly where the overlap is.
- Later rules (unevaluatedProperties and unevaluatedItems) need to know which parts of the document every matching choice inside anyOf and oneOf looked at. The checker could stop at the first matching choice to save time, but then those later rules could give a wrong answer. Which should it do? — Always check every choice — Recommended. anyOf and oneOf look at every choice so the information needed by later rules is complete and correct. Cost: a little slower on schemas with many choices.
- When a document fails 'allOf' because several of its rules fail, what should the 'basic' report list? — Every failing rule's own failures — Recommended. One entry saying allOf failed, then the reasons from every rule that failed, each with its location and rule (never the data). This matches how anyOf and oneOf are reported.
- 'not' fails when its inner rule matches, so there are no inner failures to show. What should the report say? — One entry for 'not' only — Recommended. A single entry with the location and the 'not' rule, for example 'not: document matched a schema it must not match'. No data and no inner details.
- With 'if/then/else', a failure inside 'if' alone never makes the document invalid. But when 'then' or 'else' fails, what should the 'basic' report list? — The failing branch and its own failures — Recommended. One entry for the branch that failed ('then' or 'else'), then the reasons inside it, each with its location and rule (never the data). Nothing is listed for 'if', since 'if' itself never fails a document. This matches how allOf, anyOf and oneOf are reported.
- Operator note on refining again — Add to this story (finding F2 from story 180's review): child keyword locations are currently relative to the child, because compile.Compile always starts at "" and CompileContext carries no location. The applicators here are the first real children, so add a way to compile a subschema at its schema location (e.g. a CompileAt / location field on CompileContext) so every unit's keywordLocation is a JSON Pointer from the schema root (e.g. /allOf/0/const), and update story 180's units_test.go TestUnitOrderAndNoValues assertion to the rooted value. Keep everything else in the accepted plan.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- Story 180 is 'started', not released. The exact path of units_test.go (assumed internal/output/units_test.go), the output-unit type and whether internal/output/basic.go already renders child failure units are not visible in the inventory. If basic.go needs changes to list nested failures, it must be added to the write claims.
- It is unknown whether CompileContext already exposes the enclosing schema object. 'if' needs sibling then/else; if that access is missing, T1 adds it to registry.go and compile.go.
- Whether spec_test.go's TestSpecCases glob already covers testdata/applicator is not known. A new spec_applicator_test.go is proposed to avoid editing it; if the glob is testdata/**, that file is unnecessary and its cases would run twice.
- The location change to CompileContext touches files that approved stories 181-184 also write (registry.go, compile.go). Those stories will be scheduled after or before this one, not in parallel, and any keyword they add must use the rooted location.
- Treating an empty allOf/anyOf/oneOf array as a schema error follows the 2020-12 'MUST be a non-empty array' rule ahead of meta-schema validation (story 189). This is consistent with the existing type/enum schema-error tests.

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
