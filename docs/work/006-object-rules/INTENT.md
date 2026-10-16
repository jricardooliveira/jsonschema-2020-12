# INTENT #6

- Author: Joao Oliveira
- Date: 2026-10-16
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Add the 2020-12 object keywords properties, patternProperties, additionalProperties, propertyNames, dependentSchemas, maxProperties, minProperties, required and dependentRequired. Each keyword lives in its own file and registers through the keyword registry. patternProperties uses the existing ECMA-262-to-RE2 translation step. additionalProperties applies only to properties not matched by properties or patternProperties in the same schema object. The keywords apply only to object instances. Errors report the failing instance location without quoting values. Write spec-derived positive and negative tests for every keyword.

## Operator clarifications
- A later rule, 'unevaluatedProperties', needs to know which parts of a document these object rules already checked. Should these rules record that now, as the 2020-12 rules describe? — Yes, record it now — 'properties', 'patternProperties' and 'additionalProperties' each note which property names they checked. The later rule can then use that without reopening these files.
- When a document is missing fields the schema says are required, what should the error say? — Name the missing fields — Example: 'missing required property: email'. The field names come from the schema, not from the document, so no document data is shown.
- If an object breaks several of these rules at once, for example two wrong fields and one missing field, how many errors should the report show? — Every failure — The 'basic', 'detailed' and 'verbose' reports list all three problems, so one run shows everything to fix. The yes/no 'flag' report may stop at the first failure, because it shows no details.
- Property names (the keys in a document) are also data. For example, a document might use a person's email address as a key. When a key breaks a rule, such as 'propertyNames' rejecting a badly formed key or 'additionalProperties' rejecting an unexpected one, should the report show that key? — Show the key only as part of the location — Recommended. This follows the 2020-12 rules. Each report already gives the location as a path, for example /users/alice@example.com, so the key appears there. The message text never repeats it, and values are never shown.
- When one object breaks several of these rules, the report lists every failure. Should those failures always come out in the same order, and which order? — Same order every run, following the document — Recommended. Failures appear in the order the keys appear in the document, then missing required fields in the order the schema lists them. Running the tool twice on the same files gives identical output, so scripts and tests can compare reports directly.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The repository inventory was empty. Every path is derived from the project spec conventions, not from the actual tree. Keyword file names are assumed to be lowercased keyword names (for example additionalproperties.go). The testdata area is assumed to be the vocabulary name.
- The location of the reusable ECMA-262-to-RE2 translation file from story 182 is unknown, so it is claimed only as a component read.
- It is unknown whether the skeleton's evaluation context already offers a place for per-call keyword state. If it does, the write claim on the crowded internal/eval/eval.go is not needed.
- It is unknown whether the jsonvalue object model from story 179 preserves member order. If it does not, internal/jsonvalue/decode.go (crowded) must change, and that is outside this proposal's claims.
- The coordination with story 191 (annotation collection, dropping annotations from failed subschemas) is open. This story records evaluated names regardless of outcome, and filtering by success is assumed to be 191's or 192's job.
- Detailed and verbose ordering can only be tested once story 194 exists. Here only basic (and flag) are tested.

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
