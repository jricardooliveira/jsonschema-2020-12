# Release notes — #15: unevaluatedProperties and unevaluatedItems (2020-12 unevaluated vocabulary)

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #15
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

unevaluatedProperties and unevaluatedItems keywords are now available and working correctly. They enforce that every property or array item has been evaluated by some adjacent or nested rule (properties, patternProperties, items, allOf, anyOf, oneOf, etc.) and reject those that haven't, reporting one error per uncovered child at that child's instance location. Both are self-registering plug-ins in their own files. A new AfterSiblings flag on keywords ensures they run last and see all sibling annotations before evaluating.

- internal/keywords/registry.go: Keyword struct gains AfterSiblings boolean field (default false), marking keywords that need sibling annotations
- internal/eval/eval.go: evalNode() now executes keywords in two passes (non-flagged first, then flagged), CoveredItems() helper method to summarize item annotations by instance location, extracted evalItem() function to apply one keyword
- internal/eval/order_test.go: new test verifying AfterSiblings keywords run last regardless of key order in the schema
- internal/keywords/unevaluated/unevaluatedproperties.go: registers unevaluatedProperties keyword, applies subschema to uncovered properties, one error per uncovered property at /property instance location, records covered properties on success
- internal/keywords/unevaluated/unevaluatedproperties_test.go: unit tests for one-error-per-property, sibling coverage, failed branches, nesting, non-objects, subschema application
- internal/keywords/unevaluated/unevaluateditems.go: registers unevaluatedItems keyword, applies subschema to uncovered array items, tracks coverage from prefixItems/items/contains/nested unevaluated
- internal/keywords/unevaluated/unevaluateditems_test.go: unit tests for one-error-per-item, sibling coverage sources, contains matching only, failed branches, nesting
- testdata/unevaluated/unevaluatedProperties.json: 13 test groups (127 lines total) covering properties, patternProperties, additionalProperties, allOf, anyOf, oneOf, not, if/then/else, dependentSchemas, nested unevaluatedProperties, $ref and $dynamicRef extension, each citing 2020-12 core 11.3 or 7.7
- testdata/unevaluated/unevaluatedItems.json: 12 test groups (115 lines total) covering prefixItems, items, contains, allOf, anyOf, oneOf, not, if/then/else, nested unevaluatedItems, $ref and $dynamicRef extension, each citing 2020-12 core 11.2 or 7.7

## How to verify

- Validate {"a": 1, "b": 2} against {"properties": {"a": true}, "unevaluatedProperties": false}: must fail with exactly one error at instance location /b
- Validate [1, 2, 3] against {"prefixItems": [true], "unevaluatedItems": false}: must fail with exactly two errors at /1 and /2
- Validate {"a": 1} against {"unevaluatedProperties": false, "allOf": [{"properties": {"a": true}}]} (key order varied): must pass, showing that unevaluatedProperties sees allOf's annotations regardless of key order
- Validate {"a": 1} against {"anyOf": [{"properties": {"a": true}, "required": ["missing"]}], "unevaluatedProperties": false}: must fail, showing that annotations from a failed anyOf branch do not cover
- Run go test -race ./internal/keywords/unevaluated/: all unit tests pass
- Run go test ./... to execute the spec-derived testdata cases: all 25 test groups (13 properties + 12 items) pass
- Inspect one failing validation result: error output contains no instance values, only instance location (e.g. /nickname), keyword location and rule name

## Migration

none

## Rollback

Delete internal/keywords/unevaluated/ directory and testdata/unevaluated/ directory. Revert internal/keywords/registry.go (remove AfterSiblings field from Keyword struct). Revert internal/eval/eval.go (remove two-pass loop, restore single-pass iteration over Items, delete CoveredItems method, delete evalItem helper function). Delete internal/eval/order_test.go.

## Known limitations

- The 'not' keyword may contribute annotations from its subschema even though 'not' must fail to pass; this could cause later unevaluated* keywords to incorrectly treat properties or items as covered. This is a 'not' implementation issue, not this story. The correct valid/invalid outcome is unaffected because 'not' passes only when its subschema fails, so annotations do not propagate in normal evaluation paths.
- The evaluated-items and evaluated-properties record lookups scan the entire annotation map per keyword, resulting in O(N²) behaviour on large documents with many applicators; this is not counted against evaluation-step limits (a story 191 or follow-up concern).
- Keyword locations in error units are asserted implicitly through test pass/fail but not explicitly inspected; spot-check a failing validation to verify keyword location format (e.g., /unevaluatedProperties).

## Concerns

- The 'not' keyword may leak annotations: if the 'not' keyword registers evaluated properties or items after its subschema fails, later unevaluated* keywords will incorrectly see them as covered. This is a correctness issue owned by the 'not' implementation (story 186 or later). Recommend a follow-up investigation.
- Keyword location and absolute keyword location in error units are not explicitly tested in unevaluated*_test.go files; they are only checked indirectly through pass/fail outcome. Recommend a spot-check of actual error output from one failing case.
- The fail-closed code path (returning error when ctx.Items does not implement the property/item record interface) is not exercised by any test. This would be triggered only by a bug in story 191's EvalContext. Recommend adding a mock test if paranoia is desired.
