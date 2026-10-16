# Release notes — #6: 2020-12 object keywords: properties, patternProperties, additionalProperties, propertyNames, dependentSchemas, maxProperties, minProperties, required, dependentRequired

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #6
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

Nine 2020-12 object keywords (properties, patternProperties, additionalProperties, propertyNames, dependentSchemas, maxProperties, minProperties, required, dependentRequired) are implemented as self-registering files in the correct vocabulary packages, with spec-derived testdata and deterministic sorted-order output. All acceptance criteria are met and tests pass with the race detector. A medium-severity gap exists: no evaluation step limit on object keyword processing, so large objects with many patternProperties could cause unbounded work.

- internal/eval: new evaluated.go with RecordProperties and PropertiesAt for per-evaluation property-name tracking; evaluated_items.go extended with props flag and names field to hold property entries in the item log
- internal/keywords/applicator: five new self-registering keywords (properties, patternProperties, additionalProperties, propertyNames, dependentSchemas), each in its own file with tests; properties.go documents the operator's ordering rule (sorted keywords, sorted members, schema order for missing fields)
- internal/keywords/validation: four new self-registering keywords (maxProperties, minProperties, required, dependentRequired), each in its own file with tests; missing-field messages name only the schema field, never instance data
- testdata/applicator: four spec-derived JSON test files (properties, patternProperties, additionalProperties, propertyNames, dependentSchemas)
- testdata/validation: four spec-derived JSON test files (maxProperties, minProperties, required, dependentRequired)

## How to verify

- Compile a schema with properties (e.g. {"properties":{"a":{"type":"string"}}}), validate {"a":"x"} (pass) and {"a":1} (fail)
- Compile a schema with patternProperties (e.g. {"patternProperties":{"^p":{"type":"integer"}}}), validate {"px":1} (pass) and {"px":"x"} (fail)
- Compile a schema with additionalProperties:false and properties:{"a":{}} and patternProperties:{"^p":{}}, validate {"a":1,"px":2} (pass) and {"a":1,"other":2} (fail)
- Compile a schema with required:["x","y"], validate {"x":1,"y":2} (pass) and {"x":1} (fail); check error message contains 'missing required property: y'
- Compile a schema with propertyNames:{"maxLength":3}, validate {"a":1} (pass) and {"abcdef":1} (fail); verify the failing key 'abcdef' appears only in the instanceLocation as /abcdef, never in the error message
- Compile an invalid patternProperties pattern (e.g. {"patternProperties":{"(?=a)":true}}); expect a compile error (exit code 2), not an invalid result
- Run the same schema and document twice, render basic output both times, and compare bytes to verify determinism

## Migration

none

## Rollback

Revert the 30 added and modified files (29 new files under internal/keywords/applicator, internal/keywords/validation, testdata, and internal/eval; 1 modified file evaluated_items.go). The keywords are self-registering and add no dependencies, so no registry edits are needed.

## Known limitations

- Evaluation step limits do not count object keyword work (properties, patternProperties, additionalProperties, propertyNames, dependentSchemas, required, dependentRequired, maxProperties, minProperties): a large object with many patternProperties patterns causes unbounded evaluation steps. Mitigated by RE2 running in linear time on input size and input size being capped by a separate limit, but a future story should add step counting to object keywords.
- The operator's ordering rule (sorted keyword names, then sorted member names, then schema array order for missing fields) supersedes the story's original wording ('order the keys appear in the document'). Reviewers should check against the operator's ADR decision, not the story text.
- Detailed and verbose output shapes cannot be tested until story 194 defines them; only basic and flag are tested here.
- The evaluated-property record is stored but annotation filtering for failed subschemas is deferred to stories 191 and 192; keywords record names unconditionally, and later stories will drop annotations from failures.
- patternProperties and additionalProperties each compile their own copy of sibling patternProperties regexes (about 6 lines duplicated from validation/pattern.go); a follow-up story can extract a shared CompilePattern helper.

## Concerns

- Unbounded evaluation on large objects: the tester flagged that object keyword processing (especially patternProperties with many patterns) does not count toward the evaluation step limit. Running a schema with 50 patternProperties against an object with 1000 keys causes 50,000 regex matches, unbounded by evaluation limits. RE2 is linear so it does not hang, and input size is capped by a separate limit, so this is a resource-contention issue rather than a crash risk. However, a service accepting untrusted schemas could be DOS'd by a schema with many patterns. Consider a follow-up story to add step counting to object keywords.
- The per-evaluation property record (evaluated_items.go) is outside the original story scope but required to avoid leaking annotations on failed subschemas. The change is small (three fields in itemEntry), and the file is not crowded, but verify in code review that the addition is minimal.
- Story 191 (annotation collection and filtering) will define its own mechanism. It should build on the EvaluatedItems log (evaluated.go and evaluated_items.go) rather than duplicate state, to avoid data loss when subschemas fail.
