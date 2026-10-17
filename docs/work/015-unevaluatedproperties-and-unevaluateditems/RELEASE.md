# Release notes — #15: unevaluatedProperties and unevaluatedItems (2020-12 unevaluated vocabulary)

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #15
- Status: proposed (pending Checkpoint 3 sign-off)

## What changed

Schemas can now reject object properties and array items that no keyword covered, using unevaluatedProperties and unevaluatedItems. Each keyword reads evaluated-location annotations from every sibling applicator—properties, patternProperties, additionalProperties, prefixItems, items, contains, allOf, anyOf, oneOf, not, if/then/else, dependentSchemas, $ref, $dynamicRef and nested unevaluated keywords—and applies its subschema only to uncovered members. Failed branches and failed 'not' subschemas contribute nothing. Both keywords live in separate files, register themselves at init time, and emit one error per uncovered property or item at its instance location.

- internal/keywords/registry.go: added Scope int field to EvalContext (tracks where this schema object's evaluated-locations record began), added AfterSiblings bool flag to Keyword (marks keywords that read sibling annotations)
- internal/eval/eval.go: two-pass keyword dispatch (non-AfterSiblings first, then AfterSiblings), per-schema-object mark and rollback for evaluated-location records (so failed keywords contribute nothing), added CoveredProperties() and CoveredItems() query methods to summarise annotations per instance location
- internal/eval/order_test.go: new test verifying AfterSiblings keywords run last regardless of schema key order
- internal/keywords/unevaluated/unevaluatedproperties.go: new keyword registered to 2020-12 unevaluated vocabulary, compiles subschema, reads covered property names from current schema object's scope, applies subschema only to uncovered properties, records covered properties on success, escapes property names for JSON Pointer
- internal/keywords/unevaluated/unevaluatedproperties_test.go: comprehensive unit tests covering scoping to schema object (allOf sibling branches don't share), failed 'not' contributes nothing, error locations through $ref, nested unevaluatedProperties, failed anyOf/oneOf branches, non-object instances ignored, no instance values in output
- internal/keywords/unevaluated/unevaluateditems.go: new keyword registered to 2020-12 unevaluated vocabulary, reads covered item indices from current scope, applies subschema only to uncovered items, records covered items on success
- internal/keywords/unevaluated/unevaluateditems_test.go: unit tests mirroring properties tests for arrays (prefixItems indices, items all-items coverage, contains matching indices only)
- testdata/unevaluated/unevaluatedProperties.json: 16 spec-derived test cases each citing 2020-12 core 11.3 or 7.7 section; covers properties, patternProperties, additionalProperties, dependentSchemas, allOf, anyOf, oneOf, not, if/then/else, nested unevaluatedProperties, $ref targets, and $dynamicRef/$dynamicAnchor extensions
- testdata/unevaluated/unevaluatedItems.json: 15 spec-derived test cases each citing 2020-12 core 11.2 or 7.7; covers prefixItems, items, contains, allOf, anyOf, oneOf, not, if/then/else, nested unevaluatedItems, $ref, and $dynamicRef/$dynamicAnchor

## How to verify

- go test -race ./... passes, including internal/eval/order_test.go and both unevaluated test files
- Compile a schema with unevaluatedProperties: false and validate {"a":1,"b":2} where only "a" is covered by properties: it reports two errors at /a and /b (or just /b if "a" is covered), never leaking the values 1 or 2
- Try allOf: [{properties: {a:{}}}, {unevaluatedProperties: false}] on {a:1}: unevaluatedProperties in the second branch sees no adjacent properties keyword (only from the first branch's output), so it rejects "a" as unevaluated
- Try unevaluatedItems with a schema that has prefixItems in a sibling allOf branch: items not covered by the sibling are unevaluated (scoping rule, not shared across allOf branches)
- Compile a $ref to a schema with unevaluatedProperties: false, with local properties adjacent to the $ref: both local and referenced properties count as covered
- Validate a document with more items/properties than any rule covers: exactly one error per extra member at its own instance location (/1, /2, /age, /nickname etc.), no errors without instance location

## Migration

none

## Rollback

No persistent state or schema format change. To rollback, redeploy the previous validator binary. Schemas without unevaluatedProperties or unevaluatedItems are unaffected. Schemas that do use these keywords will fail to compile on the previous binary and need to be updated (remove the keywords) if the old binary must be kept.

## Known limitations

- unevaluatedProperties and unevaluatedItems are evaluated after all other keywords of the same schema object, whatever the key order in the schema text (by design per spec 11.2/11.3)
- Non-object instances are ignored by unevaluatedProperties; non-array instances are ignored by unevaluatedItems (per spec)
- An out-of-range Scope value (mark > len(entries)) is treated as 'no evaluated locations' rather than returning an error; this is defensive but tolerable since eval.go manages Scope
- The per-schema-object annotation lookup is a linear scan from the mark forward, not indexed; acceptable for typical nesting depth but may slow on very deep nesting (noted by QA)
- No explicit test for these keywords when output-unit or report-bytes caps are exceeded; the keywords count steps correctly and will respect limits, but edge cases at the boundary are untested

## Concerns

- These keywords depend on story 191 to provide annotation tracking (EvaluatedItems, mark/rollback) and story 191 dropping annotations from failed subschemas. The diff uses the interfaces correctly, but if 191 does not roll back on subschema failure, unevaluated* will over-report covered members. Confirm story 191 is complete and has rollback implemented.
- anyOf and oneOf must evaluate all branches to collect complete annotations (per core 10.3). If story 185 short-circuits after the first passing branch, later branches' annotations will be missing. Verify story 185 does not short-circuit.
- The $dynamicRef test cases depend on story 188 propagating annotations from dynamically resolved targets; confirm story 188 is in place.
- QA noted three low findings (out-of-range mark handling, caps testing gaps, linear scan performance): these are not blocking, but reviewers should know these improvements could be made in follow-up stories (mark validation stricter, output limits tested, indexed lookup on deep nesting).
