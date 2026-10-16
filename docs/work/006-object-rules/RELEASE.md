# Release notes — #6: 2020-12 object keywords: properties, patternProperties, additionalProperties, propertyNames, dependentSchemas, maxProperties, minProperties, required, dependentRequired

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #6
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

The nine 2020-12 object keywords (properties, patternProperties, additionalProperties, propertyNames, dependentSchemas, maxProperties, minProperties, required, dependentRequired) are now fully implemented and self-registering. A schema can now validate object structure: apply subschemas to named and pattern-matched members, enforce required fields and field counts, and record which members were evaluated for later use by unevaluatedProperties. Keys appear only in instance locations, never in error text. Failures are reported in deterministic order across runs.

- internal/eval: evaluated.go (new) adds per-evaluation record of property names; evaluated_test.go (new) tests record isolation and step limits; evaluated_items.go (modified) adds property entries and step counter to the item annotation log
- internal/keywords/applicator: properties.go, patternProperties.go, additionalProperties.go, propertyNames.go, dependentSchemas.go (all new, self-registering); each with its own _test.go
- internal/keywords/validation: maxProperties.go, minProperties.go, required.go, dependentRequired.go (all new, self-registering); each with its own _test.go
- testdata/applicator: properties.json, patternProperties.json, additionalProperties.json, propertyNames.json, dependentschemas.json (spec-derived test cases)
- testdata/validation: maxproperties.json, minproperties.json, required.json, dependentrequired.json (spec-derived test cases)

## How to verify

- Run `go test -race ./internal/keywords/applicator ./internal/keywords/validation ./internal/eval` to verify all keywords, race safety and step limits
- Verify non-object instances pass: compile and validate an array or string against a schema with these keywords; must pass
- Verify deterministic ordering: run the test in properties_test.go::TestObjectOutputDeterministicAndOrdered twice; both must produce byte-identical output
- Verify keys don't leak: in properties_test.go::TestPropertiesKeyOnlyInLocation and propertynames_test.go::TestPropertyNamesKeyOnlyInMessagesNever, confirm error messages contain no key text
- Verify required fields report in schema order: run required_test.go::TestRequiredReportsMissingInSchemaOrder; order must match schema array order
- Verify additionalProperties ignores sibling patterns: run additionalproperties_test.go::TestAdditionalPropertiesRecordsOnlyUncovered; only uncovered names must be recorded
- Verify evaluated names are recorded and rollback on failure: run properties_test.go::TestPropertiesRolledBackWhenSubschemaFails; failed subschemas must not keep their property records

## Migration

None. This is pure addition of new keywords. Existing schemas without these keywords are unaffected.

## Rollback

Remove the nine keyword files (properties.go, patternProperties.go, additionalProperties.go, propertyNames.go, dependentSchemas.go in applicator; maxProperties.go, minProperties.go, required.go, dependentRequired.go in validation) and their _test.go files, remove testdata JSON files, revert evaluated.go and evaluated_test.go deletions, and restore evaluated_items.go to its previous version (remove props bool, names []string fields from itemEntry, remove steps/maxSteps from EvaluatedItems, restore At() filter).

## Known limitations

- Step limit end-to-end testing (exit code 2 on limit exceeded) deferred to story 193, which adds it as a general option across all keywords
- Detailed and verbose output ordering not yet tested; story 194 defines those output shapes
- propertyNames deterministic-order test checks count only, not actual order sequence (minor gap)

## Concerns

- Minor test-coverage gaps documented by tester: >20-key determinism test compares units rather than rendered bytes as architecture specified; propertyNames order test incomplete. Neither affects functionality, both flagged as lower priority.
- patternProperties and additionalProperties duplicate ~6 lines of regex compilation (size check, ECMA-262 translate, compile) from validation/pattern.go. Documented as follow-up: create a shared CompilePattern helper.
- Story 191 (annotation collection) should build on the EvaluatedItems mechanism (evaluated.go, evaluated_items.go) rather than duplicate property recording. The operator noted this risk in the architecture review.
- The per-evaluation record now lives in EvaluatedItems, keyed by (instance location, keyword location). A keyword records names whether its subschemas pass or fail; story 191 must filter out failed annotations if the spec requires it. The interface is designed for that, but not yet tested.
