# INTENT #7

- Author: Joao Oliveira
- Date: 2026-10-16
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Add the 2020-12 array keywords prefixItems, items, contains, maxContains, minContains, maxItems, minItems and uniqueItems. Each keyword lives in its own file and registers through the keyword registry. 'items' applies only to elements after those covered by prefixItems. maxContains and minContains work with contains, and minContains 0 makes contains always pass. uniqueItems uses deep JSON equality, so 1 and 1.0 are duplicates and objects compare regardless of key order. The keywords apply only to array instances. Write spec-derived positive and negative tests for every keyword.

## Operator clarifications
- Some of these array rules also leave notes for a later rule, 'unevaluatedItems', which checks the array elements no other rule looked at. For example, prefixItems notes 'the first 3 elements were checked', and contains notes which elements matched. Should this story also record those notes, or only give the valid/invalid answer? — Record the notes now — Recommended. prefixItems, items and contains record which elements they checked, exactly as 2020-12 describes, with tests for it. The later unevaluatedItems story can then rely on them without changing these files. Cost: contains must look at every element instead of stopping at the first match.
- uniqueItems compares elements with each other to find duplicates. On a very large array, for example 100,000 items from an untrusted file, comparing every pair could take minutes. How should that be handled? — Fast duplicate check, tested on a large array — Recommended. Elements are grouped by a fingerprint first, so only likely duplicates are compared in full, and 1 and 1.0 still count as duplicates. A test checks that a large array is handled quickly. Cost: a little more code in this one keyword.
- When several elements of an array break a rule, for example items requires numbers and elements 2, 5 and 9 are text, what should the report list? — Every failing element — Recommended. The report lists each failing position (2, 5 and 9) with the rule that failed, never the value itself. This matches the 'basic' report shape, and a user can fix everything in one pass. Cost: the check always looks at the whole array, even when the answer is already 'invalid'.
- When uniqueItems finds duplicates, what should the report say? For example, elements 1 and 4 are both {"a": 1}. — Name the positions of the duplicates — Recommended. For example 'elements 1 and 4 are equal', giving only the positions and never the values, so the report stays safe for logs and still points straight at the problem.
- Sometimes an array has many copies of the same element, for example the same {"a": 1} at 500 different positions. How should the uniqueItems report list them? — One entry per group of equal elements — Recommended. Each group of equal elements gets one entry listing all its positions, for example 'elements 1, 4 and 9 are equal'. The report stays short even for a large hostile file, and it gives positions only, never values.
- The fast duplicate check gets a test on a large array. What should that test require so you know it is fast enough? — No wall-clock threshold: timing tests are flaky on a busy machine and under the race detector. Make the test deterministic instead: on 100,000 mixed elements (numbers, text, objects, with 1 and 1.0 as duplicates) it checks the correct groups and that the number of full deep-equality comparisons stays near-linear (for example far below n*n/100), using a comparison counter visible to tests.
- When 'contains' fails, no element (or too few) matches the rule. For example, 'contains' asks for at least one number and all 100,000 elements are text. What should the report list? — One entry for the array — Recommended. One entry at the array's position says 'contains' failed, with how many elements matched and how many were required. It shows no values. The report stays short even for a huge hostile file. Elements that do not match are expected for 'contains', so they are not errors themselves.
- Operator note on refining again — Build this story BEFORE story 191 and remove the dependency on 191. Record the array annotations (prefixItems: largest index applied or true; items: true; contains: matching indices, or true when every element matched, as 2020-12 core 10.3.1 says) in the per-evaluation record of evaluated locations, kept in the evaluation context and never in the compiled schema, the same mechanism story 183 adds for properties (internal/eval/evaluated.go). If that file exists when this story is built, extend it; if not, add the item half in a new file internal/eval/evaluated_items.go with the same style. Story 191 later builds general annotation collection on top of these records and must not rewrite these keyword files; story 192 reads them. Other open points are accepted as stated.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory is empty, so no repository file was seen. Vocabulary package names (assumed internal/keywords/applicator and internal/keywords/validation), the registry signature, sibling-keyword access at compile time, the error type and the shared testdata runner location are all assumed from the spec and must be checked against the skeleton from story 178.
- items needs its sibling prefixItems length, and contains needs its sibling minContains and maxContains. If the story 178 registry gives compile functions no access to sibling keywords, a registry extension is needed. That is outside this story's file claims and should be raised rather than worked around.
- Per the operator note: if story 183 has already created internal/eval/evaluated.go, the items record extends it. Otherwise only evaluated_items.go is added. Both are claimed for writing, so this story and 183 are not built at the same time. Story 183 is not a dependency.
- Attaching the record to the evaluation context probably needs one line in the crowded file internal/eval/eval.go, unless evaluated.go from story 183 already provides a slot. The claim is kept so the story does not look independent of eval.go.
- The error-report criteria are tested at the Result or error level in the library. Tests of the rendered 'basic' JSON shape would need story 180. It is not listed as a dependency because the instance location and keyword location are expected to come from the skeleton's error model.
- Boolean schemas as subschemas of prefixItems, items and contains are assumed to be compiled by the skeleton. Story 185 states this for every position, so if the skeleton handles only a root boolean, story 185 or a small skeleton fix is needed first.
- The exact wording of the 2020-12 contains annotation (indices, or true when every element matched) follows the operator's reading of core 10.3.1.3. Test citations should quote that section.

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
