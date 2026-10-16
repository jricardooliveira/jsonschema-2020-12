# SPEC #8: Applicator keywords allOf, anyOf, oneOf, not and if/then/else, with keyword locations rooted at the schema root

- Date: 2026-10-16
- Type: feature
- Verdict: pass
- Status: proposed (pending Checkpoint 1 sign-off)

## Problem

The validator cannot yet combine subschemas. Without allOf, anyOf, oneOf, not and if/then/else, most real 2020-12 schemas cannot be checked. Child keyword locations are also currently relative to the child (finding F2 from story 180's review): compile.Compile always starts at "" and CompileContext carries no location, so a failure inside allOf/0 would report /const instead of /allOf/0/const.

## Why this matters

These applicators are mandatory 2020-12 behaviour, listed in the brief's 'Mandatory scope'. They are also the first keywords with real child subschemas, so this is where keyword locations must become JSON Pointers from the schema root. Without that, every later applicator (properties, items, $ref) and every 'basic' report would point at the wrong rule.

## Acceptance criteria

- Subschemas are compiled at their schema location: CompileContext carries the current keyword location and offers a way to compile a child at a location (for example CompileAt). Every output unit's keywordLocation is a JSON Pointer from the schema root, for example /allOf/0/const. Story 180's units_test.go TestUnitOrderAndNoValues asserts the rooted value.
- allOf (internal/keywords/applicator/allof.go) is valid only when every subschema is valid. On failure, the basic report has one allOf entry followed by the failures of every failing subschema. Each entry gives the instance location and the keyword location, and never the data.
- anyOf (anyof.go) is valid when at least one subschema is valid. When none match, the basic report has one anyOf entry followed by each choice's own failures, each with its location and rule and never the data.
- oneOf (oneof.go) is valid only when exactly one subschema is valid. When zero choices match, it reports like anyOf. When more than one matches, its single entry names the matching choice indices, for example 'oneOf: choices 0 and 2 both matched', with location and rule but no data.
- anyOf and oneOf always evaluate every choice and never stop at the first match. A test proves that a later choice is still evaluated after an earlier one matched.
- not (not.go) is valid only when its subschema is invalid. On failure it reports exactly one entry with the location and the 'not' rule (for example 'not: document matched a schema it must not match') and no inner details or data.
- if/then/else (if.go, then.go, else.go): 'then' and 'else' have no effect without 'if'. A failure inside 'if' alone never makes the document invalid. When 'if' passes, 'then' applies; when it fails, 'else' applies. A failing branch reports one entry for that branch followed by its own failures, and nothing is listed for 'if'.
- Each keyword lives in its own file and registers itself through the keyword registry from init(); no central keyword list or switch is edited. A malformed value is a compile (schema) error, never 'invalid': allOf/anyOf/oneOf that is not a non-empty array of schemas, or not/if/then/else that is not a schema.
- The boolean schemas true and false work as subschemas in every one of these keywords.
- Spec-derived positive and negative JSON cases, each citing the 2020-12 core section (10.2.1 / 10.2.2), cover every keyword. They include nested combinations (for example allOf inside anyOf inside not) and oneOf with more than one match. Report tests assert that no instance value appears in the output.

## Non-goals

- Annotation collection and evaluated-location tracking for unevaluatedProperties/unevaluatedItems (stories 191, 192). This story only guarantees that every anyOf/oneOf choice is evaluated.
- Object and array applicators (stories 183, 184) and $ref/$dynamicRef (stories 186, 188).
- Meta-schema validation of schemas (story 189).
- detailed/verbose output shapes (story 194).
- Limits on nesting or recursion (story 193).

## Tasks

- **T1** — Root keyword locations: compile and evaluate subschemas at their schema location
  - purpose: Give CompileContext the current keyword location and a CompileAt-style child compile, and give EvalContext a way to evaluate a compiled child and get its result and failure units. Child units then carry rooted keyword locations. Update story 180's TestUnitOrderAndNoValues to the rooted value.
  - scope: internal/keywords/registry.go, internal/compile/compile.go, internal/compile/compile_location_test.go, internal/eval/eval.go, internal/eval/eval_subschema_test.go, internal/output/units_test.go
  - done when: go test -race ./internal/... passes. A child const inside allOf/0 reports keywordLocation /allOf/0/const. TestUnitOrderAndNoValues asserts the rooted location.
  - delivers: AC1
- **T2** — allOf and not
  - purpose: Add allof.go and not.go, plus a shared helper for compiling a schema-array value at /<keyword>/<i>, with the nested-failure reporting and schema-error checks.
  - scope: internal/keywords/applicator/subschemas.go, internal/keywords/applicator/allof.go, internal/keywords/applicator/allof_test.go, internal/keywords/applicator/not.go, internal/keywords/applicator/not_test.go
  - done when: allof_test.go and not_test.go pass, covering valid, invalid, boolean subschemas, report entries without values, and schema errors.
  - depends on: T1
  - delivers: AC2, AC6, AC8, AC9
- **T3** — anyOf and oneOf
  - purpose: Add anyof.go and oneof.go. Both evaluate every choice. oneOf names the matching indices when more than one choice matches. Both report each choice's failures when none match.
  - scope: internal/keywords/applicator/anyof.go, internal/keywords/applicator/anyof_test.go, internal/keywords/applicator/oneof.go, internal/keywords/applicator/oneof_test.go
  - done when: anyof_test.go and oneof_test.go pass, including the every-choice-evaluated test and the 'choices 0 and 2 both matched' message.
  - depends on: T2
  - delivers: AC3, AC4, AC5, AC8, AC9
- **T4** — if/then/else
  - purpose: Add if.go, then.go and else.go. 'if' evaluates its subschema without reporting it and then applies the sibling 'then' or 'else'. 'then' and 'else' register with compile-only behaviour, so they do nothing alone.
  - scope: internal/keywords/applicator/if.go, internal/keywords/applicator/if_test.go, internal/keywords/applicator/then.go, internal/keywords/applicator/then_test.go, internal/keywords/applicator/else.go, internal/keywords/applicator/else_test.go
  - done when: if/then/else tests pass: then/else ignored without if, if failure alone is valid, and a failing branch reports the branch entry plus its failures and nothing for if.
  - depends on: T2
  - delivers: AC7, AC8, AC9
- **T5** — Spec-derived applicator cases
  - purpose: Add testdata JSON cases with spec citations for every keyword, boolean subschemas, nested combinations and multi-match oneOf, run through the shared speccases runner from a new test file.
  - scope: testdata/applicator/allOf.json, testdata/applicator/anyOf.json, testdata/applicator/oneOf.json, testdata/applicator/not.json, testdata/applicator/if-then-else.json, testdata/applicator/boolean-subschemas.json, testdata/applicator/nested.json, spec_applicator_test.go
  - done when: go test ./... runs every testdata/applicator case, each with a citation; all pass.
  - depends on: T3, T4
  - delivers: AC2, AC3, AC4, AC6, AC7, AC9, AC10
